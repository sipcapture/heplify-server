Homer 5, heplify-server, TICK Stack
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

## Notes
When dealing with prometheus counters in InfluxDB, refer to the following differential example:
```
SELECT  difference(last("counter")) AS "mean_counter" FROM "homer"."autogen"."heplify_method_response" WHERE time > :dashboardTime: GROUP BY time(:interval:), "method", "response" FILL(null)
```

![image](https://user-images.githubusercontent.com/1423657/40862016-705d998a-65eb-11e8-8b03-e711b7b4498d.png)
