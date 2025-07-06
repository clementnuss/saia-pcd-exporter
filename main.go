package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/VictoriaMetrics/metrics"
	"github.com/clementnuss/saia-grpc-service/gen/go/saia/v1/saiav1connect"
	"github.com/clementnuss/saia-pcd-exporter/internal"
	"golang.org/x/net/http2"
)

// caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
var (
	csvFile    = flag.String("csv_file", "misc/prometheus_metrics.csv", "The file containing the metric list")
	serverAddr = flag.String("addr", "http://192.168.85.41:50051", "The server address in the format of host:port")
)

// serverHostOverride = flag.String("server_host_override", "x.test.example.com", "The server name used to verify the hostname returned by the TLS handshake")

func main() {
	flag.Parse()

	client := saiav1connect.NewSaiaPcdServiceClient(
		&http.Client{
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			},
		},
		*serverAddr,
		connect.WithGRPC(),
	)

	e, err := internal.NewBiogasExporter(*csvFile, client)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}
	// Start metric collection (update every 30 seconds)
	e.Start(30 * time.Second)
	defer e.Stop()

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, false)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"metrics_count": %d, "metrics": [`, len(e.Metrics))
		for i, metric := range e.Metrics {
			if i > 0 {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `{"name": "%s", "type": "%v", "address": "%d", "description": "%s"}`,
				metric.Name, metric.RegisterType, metric.Address, metric.Description)
		}
		fmt.Fprint(w, "]}")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
