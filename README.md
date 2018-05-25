![image](https://user-images.githubusercontent.com/1423657/38167610-1bccc596-3538-11e8-944c-8bd9ee0433b2.png)

**heplify-server** is a stand-alone **HOMER** *Capture Server* developed in Go, optimized for speed and simplicity. Distributed as a single binary ready to capture TLS and UDP **HEP**, Protobuf encapsulated packets from [heplify](https://github.com/sipcapture/heplify) or any other [HEP](https://github.com/sipcapture/hep) enabled agent, indexing to database and rotating using H5 or H7 table format. **heplify-server** provides precise SIP and RTCP metrics with the help of Prometheus and Grafana. It gives you the possibility to get a global view on your network and individual SIP trunk monitoring.

*TLDR; minimal, stand-alone HOMER capture server without Kamailio or OpenSIPS dependency. It's not as customizeable as Kamailio or OpenSIPS with their configuration language, the focus is simplicity!*


### Installation
You have 3 options to get **heplify-server** up and running:

* Download a [release](https://github.com/sipcapture/heplify-server/releases)
* Docker [compose](https://github.com/sipcapture/heplify-server/tree/master/docker/hom5-hep-prom-graf)
* Compile from sources:  
  
  [install](https://golang.org/doc/install) Go > 1.10  
  cd cmd/heplify-server  
  go get -u -v -d  
  go build  


### Configuration
heplify-server can be configured using command-line options, or by defining a local [configuration file](https://github.com/sipcapture/heplify-server/blob/master/example/)

------

### Testing
##### Stand-Alone
```
./heplify-server -h
```
##### Docker
A sample Docker [compose](https://github.com/sipcapture/heplify-server/tree/master/docker/hom5-hep-prom-graf) file is available providing heplify-server, Homer 5 UI, Prometheus, Alertmanager and Grafana in seconds!
```
cd heplify-server/docker/hom5-hep-prom-graf/
docker-compose up -d
```
##### Service
A sample service file is available under `/example`
```
cp example/heplify-server.service /etc/systemd/system/
systemctl daemon-reload
systemctl start heplify-server
systemctl enable heplify-server
```

### Support
* Testers, Reporters and Contributors [welcome](https://github.com/sipcapture/heplify-server/issues)

### Screenshots
![sip_metrics](https://user-images.githubusercontent.com/20154956/39880524-57838c04-547e-11e8-8dec-262184192742.png)
![xrtp](https://user-images.githubusercontent.com/20154956/39880861-4b1a2b34-547f-11e8-8d38-69fa88713aa9.png)
----
#### Made by Humans
This Open-Source project is made possible by actual Humans without corporate sponsors, angels or patreons.<br>
If you use this software in production, please consider supporting its development with contributions or [donations](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=donation%40sipcapture%2eorg&lc=US&item_name=SIPCAPTURE&no_note=0&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donateCC_LG%2egif%3aNonHostedGuest)

[![Donate](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_donations&business=donation%40sipcapture%2eorg&lc=US&item_name=SIPCAPTURE&no_note=0&currency_code=EUR&bn=PP%2dDonationsBF%3abtn_donateCC_LG%2egif%3aNonHostedGuest) 
