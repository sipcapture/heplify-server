module github.com/sipcapture/heplify-server

go 1.14

require (
	github.com/VictoriaMetrics/fastcache v1.5.7
	github.com/antonmedv/expr v1.8.8
	github.com/aws/aws-sdk-go-v2 v1.16.2
	github.com/aws/aws-sdk-go-v2/config v1.15.3
	github.com/aws/aws-sdk-go-v2/credentials v1.11.2
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.26.3
	github.com/buger/jsonparser v1.1.1
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/gobwas/ws v1.0.3
	github.com/gogo/protobuf v1.3.2
	github.com/golang/snappy v0.0.3
	github.com/lib/pq v1.10.4
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/negbie/cert v0.0.0-20190324145947-d1018a8fb00f
	github.com/negbie/logp v0.0.0-20190313141056-04cebff7f846
	github.com/negbie/multiconfig v1.0.0
	github.com/olivere/elastic v6.2.33+incompatible
	github.com/pelletier/go-toml v1.8.0
	github.com/prometheus/client_golang v1.7.0
	github.com/prometheus/common v0.10.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sipcapture/golua v0.0.0-20200610090950-538d24098d76
	github.com/stretchr/testify v1.7.1
	github.com/valyala/bytebufferpool v1.0.0
	github.com/valyala/fasttemplate v1.1.1
	github.com/xitongsys/parquet-go v1.6.2
	github.com/xitongsys/parquet-go-source v0.0.0-20221025031416-9877e685ef65
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.45.0
)
