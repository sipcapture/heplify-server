package database

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/xitongsys/parquet-go-source/s3v2"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"

	"golang.org/x/sync/syncmap"
)

type Parquet struct {
	db         *sync.Map
	s3hash     map[string]*S3BucketHash
	bulkCnt    int
	globalLock sync.RWMutex
	index      int
}

// HEP represents HEP packet
type S3BucketHash struct {
	PartName      string
	S3File        source.ParquetFile
	ParquetWriter *writer.ParquetWriter
	Error         error
	CallCnt       int
	Index         int
	Active        bool
	LastUsage     time.Time
}

func (m *Parquet) setup() error {
	m.db = new(syncmap.Map)
	m.s3hash = map[string]*S3BucketHash{}
	m.bulkCnt = config.Setting.DBBulk
	go m.doCheckTimeoutCheck()

	return nil
}

func (m *Parquet) insert(hCh chan *decoder.HEP) {

	ctx := context.Background()

	//Check if we use AWS or Cloudflare (DBSchema = s3 or r2 )
	if config.Setting.DBAddr != "" {

		if !strings.HasPrefix(strings.ToLower(config.Setting.DBAddr), "http") {
			logp.Err("you have to provide the DBAddr that starts with https:// or http://")
		}
	}

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: config.Setting.DBAddr,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithEndpointResolverWithOptions(r2Resolver),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.Setting.DBUser, config.Setting.DBPass, "")),
	)
	if err != nil {
		logp.Err("couldn't load aws config ", "%v", err)
		return
	}

	//generate own s3client
	s3client := s3.NewFromConfig(cfg)

	// this for later, which we can overwrite
	/* uploaderOptions := make([]func(*manager.Uploader), 0)
	uploader_cloudflare := func(param *manager.Uploader) {
		param.PartSize = 5 * 1024 * 1024 // 5MB per part
		param.LeavePartsOnError = false
	}
	uploaderOptions = append(uploaderOptions, uploader_cloudflare)
	*/

	parq := decoder.HEPParquet{}
	var partTime string

	for pkt := range hCh {
		//date := pkt.Timestamp.Format(time.RFC3339Nano)
		logp.Debug("read:", " pkt %v ", pkt)

		parq.Version = int32(pkt.Version)
		parq.Protocol = int32(pkt.Protocol)
		parq.SrcIP = pkt.SrcIP
		parq.DstIP = pkt.DstIP
		parq.SrcPort = int32(pkt.SrcPort)
		parq.DstPort = int32(pkt.DstPort)
		parq.Tsec = int64(pkt.Tsec)
		parq.Tmsec = int64(pkt.Tmsec)
		parq.ProtoType = int32(pkt.ProtoType)
		parq.NodeID = int32(pkt.NodeID)
		parq.NodePW = pkt.NodePW
		parq.Payload = pkt.Payload
		parq.CID = pkt.CID
		parq.Vlan = int32(pkt.Vlan)
		parq.ProtoString = pkt.ProtoString

		unixTimeUTC := time.Unix(parq.Tsec, parq.Tmsec) //gives unix time stamp in utc

		partTime = fmt.Sprintf("%d/%02d/%02d/%02d/%01d0/hep_proto_%d", unixTimeUTC.Year(), unixTimeUTC.Month(), unixTimeUTC.Day(), unixTimeUTC.Hour(), int(unixTimeUTC.Minute()/10), parq.ProtoType)

		//Check if we have partTime and create it!
		if _, ok := m.s3hash[partTime]; !ok {

			_, err := m.recreateRecord(ctx, s3client, partTime, 0)
			if err != nil {
				logp.Debug("Couldn't create a new s3bucketg:", "partTime:%s, Error:%v", partTime, err)
				return
			}
		}

		//Write data to bucket
		m.globalLock.Lock()

		m.s3hash[partTime].CallCnt++
		m.s3hash[partTime].LastUsage = time.Now()

		if err = m.s3hash[partTime].ParquetWriter.Write(parq); err != nil {
			logp.Err("Couldn't write to parquet file error:", ":%v", err)
			m.s3hash[partTime].Error = err
			m.s3hash[partTime].Active = false
		}

		//Bump data if it excited of bulk size
		if m.s3hash[partTime].CallCnt > m.bulkCnt {

			logp.Debug("we have some records in our bulk, lets flush it", "partTime: %s", partTime)

			if err = m.s3hash[partTime].ParquetWriter.Flush(true); err != nil {
				logp.Err("parquet flush: ", " error %v", err)
				m.s3hash[partTime].Active = false
			}
			m.s3hash[partTime].CallCnt = 0
		}

		//Check the parquet size
		if m.s3hash[partTime].ParquetWriter.Size > 0 {

			logp.Debug("we have some records in buffer, lets flush it", "partTime: %s", partTime)

			if err = m.s3hash[partTime].ParquetWriter.Flush(true); err != nil {
				logp.Err("Flush error for parquet file : ", "partime: %s,  error %v", partTime, err)
				m.s3hash[partTime].Active = false
			}
		}

		m.globalLock.Unlock()

		//Check if it the s3 connection still active
		if !m.s3hash[partTime].Active || m.s3hash[partTime].ParquetWriter.Size > 0 {

			logp.Debug("bucket is not active", "we stop it: %s", partTime)

			newIndex := m.s3hash[partTime].Index
			newIndex++

			m.globalLock.Lock()

			if err = m.s3hash[partTime].ParquetWriter.WriteStop(); err != nil {
				logp.Err("bucket writeStop ", "error: %v", err)
			}

			err = m.s3hash[partTime].S3File.Close()
			if err != nil {
				logp.Err("error closing buket ", " error %v", err)
			}

			delete(m.s3hash, partTime)

			m.globalLock.Unlock()

			//Now lets recreate it
			_, err := m.recreateRecord(ctx, s3client, partTime, newIndex)
			if err != nil {
				logp.Debug("Couldn't create a new s3bucketg:", "partTime:%s, Error:%v", partTime, err)
				return
			}

			return
		}
	}
}

// make a ping keep alive
func (m *Parquet) doCheckTimeoutCheck() {

	for {
		time.Sleep(time.Duration(60) * time.Second)
		for key, element := range m.s3hash {
			//fmt.Println("Key:", key, "=>", "Element:", element)
			logp.Debug("doCheckTimeoutCheck:", "- key: %v - value: %v", key, element)

			t1 := time.Now()
			if t1.Sub(element.LastUsage).Seconds() > 300 {
				logp.Debug("doCheckTimeoutCheck - bucket is not active more than 300 seconds ", "we close it: %s", key)

				m.globalLock.Lock()

				if err := element.ParquetWriter.WriteStop(); err != nil {
					logp.Err("doCheckTimeoutCheck bucket writeStop ", "error: %v", err)
				}

				err := element.S3File.Close()
				if err != nil {
					logp.Err("doCheckTimeoutCheck error closing buket ", " error %v", err)
				}

				delete(m.s3hash, key)

				m.globalLock.Unlock()

			}
		}
	}
}

func (m *Parquet) recreateRecord(ctx context.Context, s3client *s3.Client, partTime string, index int) (*S3BucketHash, error) {

	var err error

	s3b := S3BucketHash{}
	s3b.Index = index

	//file key
	fileKey := partTime + fmt.Sprintf("_%d_%d", m.index, s3b.Index) + ".parquet"
	// create new S3 file writer
	s3b.S3File, err = s3v2.NewS3FileWriterWithClient(ctx, s3client, config.Setting.DBDataTable, fileKey, nil)
	if err != nil {
		logp.Err("Can't open file:", "filekey:%s, Error:%v", fileKey, err)
		return nil, err
	}

	//write
	s3b.ParquetWriter, err = writer.NewParquetWriterFromWriter(s3b.S3File, new(decoder.HEPParquet), 4)
	if err != nil {
		logp.Err("Can't create parquet writer", ": %v", err.Error())
		return nil, err
	}

	//SNAPPY codec and set to active
	s3b.ParquetWriter.CompressionType = parquet.CompressionCodec_SNAPPY
	s3b.Active = true

	//Lets now lock it
	m.globalLock.Lock()
	m.s3hash[partTime] = &s3b
	m.globalLock.Unlock()

	return &s3b, nil
}
