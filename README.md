A simple example of instrumenting go app using prometheus and then configuring Thanos on top of it.

---

To run prometheus

```
prometheus --config.file="prometheus.yml" --web.listen-address="0.0.0.0:9090"
```


## Some analogies

1. Instant Vector

This is basically the value of a metric at the current time.

2. Range Vector

This is the value of a metric over a range of time. Say give me the http_total_request[5m] then it will give you the recorded value of it at each timestamp it recorded (based on scrape_interval) in the last 5 minutes.

3. Scalar

This is a constant value. For example, the sum of all the values of a metric in the last 5 minutes.
