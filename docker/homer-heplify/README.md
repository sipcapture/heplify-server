![image](https://user-images.githubusercontent.com/1423657/37555893-00e3f172-29ef-11e8-9112-503f0d0f37a1.png)

## Docker Bundle
This docker compose bundle provides:
* heplify-server
  - HEP capture services
    - 9060: HEP socket
    - 9999: stats socket
* homer UI
  - HEP Search and Visualization
    - 80: HOMER UI
* mysql 5.7
  - HEP storage and indexing
    - 3306: MySQL socket
* telegraf
  - Statistics Aggregation
    - 8092: UDP Socket
* telestats
  - HEP stats aggregation Telegraf -> MySQL
    - 9999: UDP JSON socket


### Usage
Spin up a full stack for development
```
git clone https://github.com/sipcapture/heplify-server
cd heplify-server/docker/homer-heplify
```
```
docker-compose up -d
```

### Test
Test your setup with a real HEP agent or by using HEPGen
```
git clone https://github.com/sipcapture/hepgen.js
cd hepgen.js && npm install
nodejs hepgen.js -s 9060 -c ./config/b2bcall_rtcp.js
```

![ezgif com-optimize 45](https://user-images.githubusercontent.com/1423657/37555986-4b64efb6-29f0-11e8-8de3-68428da0bbb4.gif)
