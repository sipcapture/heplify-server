Homer5, heplify-server, Prometheus, Grafana Stack
========

Run from this folder  

```bash
docker-compose up
```

To bring up  

* adminer localhost:8080 (root/) empty password
* homer localhost:9080 (admin/test123) 
* prometheus localhost:9090 (admin/admin)
* alertmanager localhost:9093 (admin/admin)
* grafana localhost:3000 (admin/admin)

When the grafana dashboard autoprovisioning does not work for you make sure you have no old grafana volumes.

When you change some files inside the prometheus folder you can reload them without interruption:
```bash
curl -s -XPOST localhost:9090/-/reload -u admin:admin
```

When you change some files inside the alertmanager folder you can reload them without interruption:
```bash
curl -s -XPOST localhost:9093/-/reload -u admin:admin
```

When you need to change the docker-compose file i.e to setup smtp for grafana:
```bash
docker-compose up -d
```
Docker will only restart the service you changed inside the docker-compose file. 