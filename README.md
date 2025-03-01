A simple example of instrumenting go app using prometheus and then configuring Thanos on top of it.

---

To run prometheus

```
prometheus --config.file="prometheus.yml" --web.listen-address="0.0.0.0:9090"
```


You can then curl to `host:port/throw-random-response` endpoint with necessary json payload (which in this case is ` '{"instance_name": active_user_session}' ` (eg: '{"prod": 69420}'))

## Some analogies

1. Instant Vector

This is basically the value of a metric at the current time.

2. Range Vector

This is the value of a metric over a range of time. Say give me the http_total_request[5m] then it will give you the recorded value of it at each timestamp it recorded (based on scrape_interval) in the last 5 minutes.

3. Scalar

This is a constant value. For example, the sum of all the values of a metric in the last 5 minutes.

---

That was it about what we can do with this less metrics 😅

Next, for configuring thanos we need some HA prometheus setup + some good amount of metric. So I'll be using [this](https://github.com/thanos-io/thanos/blob/main/tutorials/interactive-example/README.md)
