package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"passkey-auth/internal/auth"
	"passkey-auth/internal/config"
	"passkey-auth/internal/cors"
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
	api.HandleFunc("/config", h.GetConfig).Methods("GET")
	api.HandleFunc("/register/begin", h.BeginRegistration).Methods("POST")
	api.HandleFunc("/register/finish", h.FinishRegistration).Methods("POST")
	api.HandleFunc("/login/begin", h.BeginLogin).Methods("POST")
	api.HandleFunc("/login/finish", h.FinishLogin).Methods("POST")
	api.HandleFunc("/logout", h.Logout).Methods("POST")
	api.HandleFunc("/auth/status", h.GetAuthStatus).Methods("GET")
	api.HandleFunc("/users", h.ListUsers).Methods("GET")
	api.HandleFunc("/users", h.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", h.UpdateUser).Methods("PUT")
	api.HandleFunc("/users/{id}", h.DeleteUser).Methods("DELETE")

	// Auth backend endpoints for ingress controllers
	// Nginx auth_request: Returns 200 for authenticated, 401 for unauthenticated
	router.HandleFunc("/auth/nginx", h.AuthCheckNginx).Methods("GET", "HEAD")
	
	// Traefik ForwardAuth: Returns 200 for authenticated, 302 redirect for unauthenticated
	router.HandleFunc("/auth/traefik", h.AuthCheckTraefik).Methods("GET", "HEAD")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}).Methods("GET")

	// Static files for admin UI
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/"))).Methods("GET")

	// Setup CORS with wildcard support
	corsHandler := cors.WildcardCORS(cors.Config{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := corsHandler(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server with proper timeouts to prevent resource exhaustion
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logrus.Infof("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
