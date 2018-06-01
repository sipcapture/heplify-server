Homer 5, heplify-server, TICK Stack
========

## Setup

```bash
docker-compose up
```

to bring up:  

* HEPlify-server localhost:9060 (hep-only)
* Homer localhost:9080 (admin/test123) 
* Chronograf localhost:9090 (admin/admin)
  * InfluxDB
  * Kapacitor
  * Telegraf

## Notes
When dealing with prometheus counters in InfluxDB, refer to the following example usage of `difference` and `derivative` functions when selecting:
```
SELECT difference(last("counter")) AS "mean_counter" FROM "homer"."autogen"."heplify_method_response" WHERE time > :dashboardTime: GROUP BY time(:interval:), "method", "response" FILL(null)
```
```
SELECT derivative(last("counter")) AS "mean_counter" FROM "homer"."autogen"."heplify_method_response" WHERE time > :dashboardTime: GROUP BY time(:interval:), "method", "response" FILL(null)
```

![image](https://user-images.githubusercontent.com/1423657/40862016-705d998a-65eb-11e8-8b03-e711b7b4498d.png)
