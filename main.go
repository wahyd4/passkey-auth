package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"

	"passkey-auth/internal/auth"
	"passkey-auth/internal/config"
	"passkey-auth/internal/database"
	"passkey-auth/internal/handlers"
)

func main() {
	// Setup logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize WebAuthn
	webAuthn, err := auth.NewWebAuthn(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn: %v", err)
	}

	// Initialize handlers
	h := handlers.New(db, webAuthn, cfg)

	// Setup routes
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/register/begin", h.BeginRegistration).Methods("POST")
	api.HandleFunc("/register/finish", h.FinishRegistration).Methods("POST")
	api.HandleFunc("/login/begin", h.BeginLogin).Methods("POST")
	api.HandleFunc("/login/finish", h.FinishLogin).Methods("POST")
	api.HandleFunc("/logout", h.Logout).Methods("POST")
	api.HandleFunc("/users", h.ListUsers).Methods("GET")
	api.HandleFunc("/users", h.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")

	// Nginx auth backend endpoint
	router.HandleFunc("/auth", h.AuthCheck).Methods("GET", "HEAD")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Static files for admin UI
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/"))).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrus.Infof("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
