![image](https://user-images.githubusercontent.com/1423657/37555893-00e3f172-29ef-11e8-9112-503f0d0f37a1.png)

## Docker Bundle
This docker compose bundle provides:
* heplify-server
  - HEP capture service
* homer UI
  - HEP Search and Visualization
* mysql 5.7
  - HEP storage and indexing
* telegraf
  - HEP stats aggregation to MySQL

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

![image](https://user-images.githubusercontent.com/1423657/37555890-f9038d1e-29ee-11e8-9dc8-ab661681a8e3.png)
