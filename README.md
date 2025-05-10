A simple example of instrumenting go app using prometheus and then configuring Thanos on top of it.

---

## (Very)Quickstart

1. 
```
go run main.go
```

2.
```
docker compose up -d
```

## Getting familiar with prometheus

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
We can run the test and it fill generate the required tsdb data for you 

- Doing `make docker` (this will be building thanos executable and also its docker image)
- Running the test (this will be running the test and generating the tsdb data)
- After this, a data dir will be created in the root of the project which will have the tsdb data'
- In that make sure to create a copy of prom1 and name it prom01 within the same directory

## Running the thanos setup

```
go run main.go # this will start the prometheus server, (just additional metrics, like histograms and native histograms)
```

```
docker compose up
```



So, now below I'll just document how things are turning out to be when setting up the thanos setup with Docker wrt to that.

---

## Some more analogies

1. StoreAPI

This is can be the GRPC endpoint that is being exposed by any of the thanos components (Storage Gateway, Sidecar, etc)

---

Please refer to `/configs` folder for all the configuration that will be used here

Now, we can create two prometheus container for our HA setup + one Prometheus (short-term) container which scrapes the metrics from their own setup endpoint

With this we can attach one Thanos sidecar to each of them (though, in this case its not exactly sidecar but in the end our sidecar will be able to expose a storeAPI endpoint for querier to query)

Further, we setup Thanos querier by providing the storeAPI GRPC endpoint of sidecar to it via `--store` flag

At the end we have a setup which is something like below

Prometheus API <- Thanos Sidecar -> Object Storage <- Storage Gateway

![image](https://github.com/user-attachments/assets/c293497c-1971-4131-9818-56a18537df88)

