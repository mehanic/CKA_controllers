// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// )

// const secretPath = "/etc/secret-volume/password"

// func getPassword() string {
// 	data, err := os.ReadFile(secretPath)
// 	if err != nil {
// 		return "❌ Failed to read secret"
// 	}
// 	return string(data)
// }

// func handler(w http.ResponseWriter, r *http.Request) {
// 	password := getPassword()
// 	fmt.Fprintf(w, "<h1>🔑 Secret Password:</h1><h2>%s</h2>", password)
// }

// func main() {
// 	http.HandleFunc("/", handler)
// 	port := "8080"
// 	log.Println("🌍 Server running on port", port)
// 	http.ListenAndServe(":"+port, nil)
// }

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const secretPath = "/etc/secret-volume/password"

// Declare a new counter metric
var passwordAccessCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "password_access_total",
		Help: "Total number of times the secret password is accessed",
	},
	[]string{"status"},
)

func init() {
	// Register the new metric with Prometheus
	prometheus.MustRegister(passwordAccessCount)
}

func getPassword() string {
	data, err := os.ReadFile(secretPath)
	if err != nil {
		// If there's an error reading the password, log it and increase the "error" counter
		passwordAccessCount.WithLabelValues("error").Inc()
		return "❌ Failed to read secret"
	}
	// Successfully read password, increase the "success" counter
	passwordAccessCount.WithLabelValues("success").Inc()
	return string(data)
}

func handler(w http.ResponseWriter, r *http.Request) {
	password := getPassword()
	fmt.Fprintf(w, "<h1>🔑 Secret Password:</h1><h2>%s</h2>", password)
}

// Metrics handler for Prometheus scraping
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the Prometheus metrics
	promhttp.Handler().ServeHTTP(w, r)
}

func main() {
	// Register the /metrics endpoint
	http.Handle("/metrics", http.HandlerFunc(metricsHandler))

	// Register the root handler for serving the secret password
	http.HandleFunc("/", handler)

	port := "8080"
	log.Println("🌍 Server running on port", port)
	// Start the HTTP server
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
