package main

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type myMetrics struct {
	total_http_requests      prometheus.Counter
	instances_active_session *prometheus.GaugeVec
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
	}

	reg.MustRegister(metrics.total_http_requests, metrics.instances_active_session)

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

	instance := make(map[string]int)

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&instance)

	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var instance_name string
	var cpu_usage int

	for k, v := range instance {
		instance_name, cpu_usage = k, v
		break // im expecting only one
	}

	metrics.instances_active_session.WithLabelValues(instance_name).Set(float64(cpu_usage))

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
