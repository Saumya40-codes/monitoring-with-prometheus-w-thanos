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

That was it about what we can do with these metrics 😅. Usually, Thanos comes into picture with good enough amount of metrics and multiple (>1) prometheus cluster scraping the metrics.

Next, for configuring thanos we need some HA prometheus setup + some good amount of metric. So I'll be using [this](https://github.com/thanos-io/thanos/blob/main/tutorials/interactive-example/README.md)


So, now below I'll just document how things are turning out to be when setting up interactive thanos setup wrt to that.

---

## Some more analogies

1. StoreAPI

This is can be the GRPC endpoint that is being exposed by any of the thanos components (Storage Gateway, Sidecar, etc)

2. Tracing

In all of the thanos components, we have specified a `--tracing.config` so we will be better be able to see the trace of a request with Jaeger

### 1. Creating Minio object store

Firstly, prometheus initially collects the metrics and aggregates most of them in memory and WAL (I just read this, have very less idea about what this will be mean) and then compacts block down to disk. Now based on that file in above link, there exists two such storage blocks, which will be stored in 2 seperate buckets in Minio

### 2. Setting up jaeger

With jaeger we can see how our the whole trace of our request lifecycle. we can specify it with `--tracing.config` flag which values point to jaeger configuration

### 3. Thanos Storage Gateways

Storage Gateway, as name suggests acts like a gateway which has idea about the tsdb blocks data stored in our minio object storage, thus querying our "old" tsdb block data


### 4. Prometheus+Thanos sidecars

- Creating two prometheus cluster in HA setup (replicas scraping same metrics)
- One prometheus cluster (short term)
- Thanos running with prometheus in each of this seperate cluster


### Querier

Now querier with the help of storeAPI now will be able to do the query task from sidecar, storage gateway based on the promql queries received


![image](https://github.com/user-attachments/assets/e877eca6-460e-4ef7-8374-62884e41cee8)

