# ![](https://i.imgur.com/QvLYJkC.png)

**heplify-server** is a stand-alone **HOMER** *v5 Capture Server* developed in GO, optimized for speed and simplicity. Distributed as a single binary ready to capture TCP/UDP **HEP** encapsulated packets, index them to database using H5 table format, produce basic usage timeseries metrics and providing users with simple, basic options for correlation and tagging inline. 

*TLDR; It's a stand-alone, minimal HOMER without Kamailio or OpenSIPS dependency/options.*

## Notice
**heplify-server** only offers a reduced set of options and is *not* designed for everyone, but should result ideal for those willing to have an *all-in-one* simple capture deployment with minimal complexity and no need for special customization.

### Status 
#### v1
* Homer 5 compatible
* Alpha Development - **NOT READY FOR PRODUCTION**
* Testers and Reporters [welcome](https://github.com/sipcapture/heplify-server/issues)
##### v2
* Homer 7 compatible
* Coming Soon

### Installation
* Download a [release](https://github.com/negbie/heplify-server/releases)
* Compile from [sources](https://github.com/negbie/heplify-server/blob/master/docker/heplify-server/Dockerfile)

### Configuration
heplify-server can be configured using command-line options, or by defining a local [configuration file](https://github.com/lmangani/heplify-server/blob/master/example/heplify-server.toml)

------

### Testing
##### Stand-Alone
```
heplify-server -h
```
##### Docker
A sample Docker stack is available providing Heplify-Server, Homer 5 UI, and basic MySQL
```
cd heplify-server/docker/homer-heplify
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
