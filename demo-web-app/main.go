package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Env       string    `json:"env"`
	Hostname  string    `json:"hostname"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
}

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "v1.0.0"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	hostname, _ := os.Hostname()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := Response{
			Message:   "Hello from Demo Web App! 🚀",
			Timestamp: time.Now(),
			Version:   version,
			Env:       env,
			Hostname:  hostname,
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		uptime := time.Since(startTime).String()
		response := HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now(),
			Uptime:    uptime,
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		info := map[string]interface{}{
			"app_name":    "demo-web-app",
			"version":     version,
			"environment": env,
			"hostname":    hostname,
			"timestamp":   time.Now(),
			"endpoints": []string{
				"/",
				"/health",
				"/api/info",
			},
		}
		json.NewEncoder(w).Encode(info)
	})

	log.Printf("🚀 Demo Web App starting on port %s", port)
	log.Printf("📦 Version: %s", version)
	log.Printf("🌍 Environment: %s", env)
	log.Printf("🖥️  Hostname: %s", hostname)
	log.Printf("🔗 Endpoints:")
	log.Printf("   GET / - Main endpoint")
	log.Printf("   GET /health - Health check")
	log.Printf("   GET /api/info - App information")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}