package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type myMetrics struct {
	total_http_requests      prometheus.Counter
	instances_active_session *prometheus.GaugeVec
	requestDuration          *prometheus.HistogramVec
	errorCount               *prometheus.CounterVec
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
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{ // okk so histogram uses buckets where for each value x in bucket accounts for value <= x in that bucket
			Name:    "request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"status_code"}),
		errorCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "request_error_counts",
			Help: "Number of request which led to an error",
		}, []string{"status_code"}),
	}

	reg.MustRegister(metrics.total_http_requests, metrics.instances_active_session, metrics.requestDuration, metrics.errorCount)

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

	timer := prometheus.NewTimer(metrics.requestDuration.WithLabelValues(fmt.Sprintf("%d", random_status_code))) // we see how interface comes into play, method NewTimer instead of looking for particular dtype or struct watches for signatures which should at minimum implement Observe method
	defer timer.ObserveDuration()

	instance := make(map[string]int)

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&instance)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var instance_name string
	var active_session int

	for k, v := range instance {
		instance_name, active_session = k, v
		break // im expecting only one
	}

	metrics.instances_active_session.WithLabelValues(instance_name).Set(float64(active_session))

	if random_status_code >= 400 {
		metrics.errorCount.WithLabelValues(fmt.Sprintf("%d", random_status_code)).Inc()
	}

	// the req-res are super fast rn, so introducing a sleep to simulate some delay
	time.Sleep((time.Duration(randIdx.Int64()) * 10) * time.Millisecond)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(random_status_code)
	w.Write([]byte("OK"))
}

func main() {
	reg := prometheus.NewRegistry()
	metrics = NewMetrics(reg)

	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{}) // custom so that no default metrics comes
	http.HandleFunc("/throw-random-response", handleRequest)
	http.Handle("/metrics", promHandler)

	http.ListenAndServe("127.0.0.1:8080", nil)
}
