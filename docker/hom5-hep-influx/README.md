Homer5, heplify-server, TICK Stack
========

## Setup

```bash
docker-compose up
```

to bring up:  

* Homer localhost:9080 (admin/test123) 
* Chronograf localhost:9090 (admin/admin)
* InfluxDB
* Kapacitor
* Telegraf
* HEPlify-server

When the Grafana dashboard autoprovisioning does not work for you make sure you have no old grafana volumes.

## Configuration

When you change some files inside the Prometheus or Alertmanager folder you can reload them without interruption.

