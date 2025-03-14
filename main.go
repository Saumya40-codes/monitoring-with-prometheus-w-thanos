package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type myMetrics struct {
	total_http_requests           prometheus.Counter
	instances_active_session      *prometheus.GaugeVec
	requestDuration               *prometheus.HistogramVec
	errorCount                    *prometheus.CounterVec
	requestDurationSecondsNative  *prometheus.HistogramVec
	requestDurationSecondsNative2 *prometheus.HistogramVec
}

var metrics *myMetrics

func NewMetrics(reg prometheus.Registerer) *myMetrics {
	metrics := &myMetrics{
		total_http_requests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_http_requests",
			Help: "Total number of http requests served",
		}),
		instances_active_session: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "instances_active_session",
			Help: "Active session of instance",
		}, []string{"instance_name"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"status_code"}),
		errorCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "request_error_counts",
			Help: "Number of request which led to an error",
		}, []string{"status_code"}),
		requestDurationSecondsNative: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:                           "request_duration_seconds_native",
			Help:                           "Time taken to serve the request, native histogram",
			NativeHistogramBucketFactor:    4, // this means our boundries will be prev_boundry * 4, // this corresponds to ratio, where schema will be schema=-1
			NativeHistogramMaxBucketNumber: 100,
		}, []string{"method", "code"}),
		requestDurationSecondsNative2: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:                           "request_duration_seconds_native2",
			Help:                           "Time taken to serve the request, native histogram",
			NativeHistogramBucketFactor:    2,   // this corresponds to ratio, where schema will be schema=0 as 2^(2^0) = 2
			NativeHistogramMaxBucketNumber: 100, // after this,it will decrease the resolution, thus schema decreases (formula is 2^(2^(-schema)))
		}, []string{"method", "code"}),
	}

	reg.MustRegister(metrics.total_http_requests, metrics.instances_active_session, metrics.requestDuration, metrics.errorCount, metrics.requestDurationSecondsNative, metrics.requestDurationSecondsNative2)

	return metrics
}

var httpstatuses []int = []int{http.StatusOK, http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusBadRequest}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	metrics.total_http_requests.Inc()

	randIdx, err := rand.Int(rand.Reader, big.NewInt(int64(len(httpstatuses))))
	if err != nil {
		panic(err)
	}
	random_status_code := httpstatuses[randIdx.Int64()]

	timer := prometheus.NewTimer(metrics.requestDuration.WithLabelValues(fmt.Sprintf("%d", random_status_code)))
	defer timer.ObserveDuration()

	timerNative := prometheus.NewTimer(metrics.requestDurationSecondsNative.WithLabelValues(r.Method, fmt.Sprintf("%d", random_status_code)))
	defer timerNative.ObserveDuration()

	timerNative2 := prometheus.NewTimer(metrics.requestDurationSecondsNative2.WithLabelValues(r.Method, fmt.Sprintf("%d", random_status_code)))
	defer timerNative2.ObserveDuration()

	if random_status_code >= 400 {
		metrics.errorCount.WithLabelValues(fmt.Sprintf("%d", random_status_code)).Inc()
	}

	randomDelay, err := rand.Int(rand.Reader, big.NewInt(1950))
	if err != nil {
		panic(err)
	}
	time.Sleep(50*time.Millisecond + time.Duration(randomDelay.Int64())*time.Millisecond)

	w.WriteHeader(random_status_code)
	w.Write([]byte("OK"))
}

func PopulateMetricsData() {
	flag := true
	devCount := 5
	stagingCount := 10
	prodCount := 15

	metrics.instances_active_session.WithLabelValues("development").Set(float64(devCount))
	metrics.instances_active_session.WithLabelValues("staging").Set(float64(stagingCount))
	metrics.instances_active_session.WithLabelValues("production").Set(float64(prodCount))

	for {

		if flag {
			devCount += 1
			stagingCount += 2
			prodCount += 1

			flag = false
		} else {
			devCount -= 2
			stagingCount -= 3
			prodCount -= 1

			flag = true
		}

		metrics.instances_active_session.WithLabelValues("development").Set(float64(devCount))
		metrics.instances_active_session.WithLabelValues("staging").Set(float64(stagingCount))
		metrics.instances_active_session.WithLabelValues("production").Set(float64(prodCount))

		resp, err := http.Get("http://127.0.0.1:8080/")
		if err != nil {
			fmt.Println("Error making request:", err)
		} else {
			resp.Body.Close()
		}

		time.Sleep(50 * time.Second)
	}
}

func main() {
	reg := prometheus.NewRegistry()
	metrics = NewMetrics(reg)

	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.HandleFunc("/", handleRequest)
	http.Handle("/metrics", promHandler)

	go http.ListenAndServe("0.0.0.0:8080", nil)

	time.Sleep(5 * time.Second)

	PopulateMetricsData()
}
