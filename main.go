package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"os"
	"sync/atomic"

	"github.com/gorilla/mux"

	"github.com/jmacneill66/go_projects/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	platform       string
	JWTSecret      string
	PolkaKey       string
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request received: %s %s\n", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		fmt.Printf("Request completed: %s %s\n", r.Method, r.URL.Path)
	})
}

func main() {

	// Load environment variables
	err1 := godotenv.Load()
	if err1 != nil {
		log.Fatal("Error loading .env file")
	}
	// Log the platform from the .env file
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM environment variable not set")
	}
	// Get DB URL from environment
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable not set")
	}
	// Open database connection
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	dbQueries := database.New(db)

	// Construct APIConfig
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		DB:             dbQueries,
		platform:       "Chirpy Platform",
		JWTSecret:      os.Getenv("JWT_SECRET"),
		PolkaKey:       os.Getenv("POLKA_KEY"),
	}

	// Set up the HTTP server
	// Create a new router
	router := mux.NewRouter() // Use gorilla/mux for routing

	// Serve static files under /app/
	fileServer := http.FileServer(http.Dir("./app"))
	router.PathPrefix("/app/").Handler(http.StripPrefix("/app/", fileServer))

	//HTTP routes / handler registrations:
	// Health check
	router.HandleFunc("/api/healthz", healthCheckHandler).Methods("GET")
	
	// Chirps
	router.HandleFunc("/api/chirps", apiCfg.createChirpHandler).Methods("POST")
	router.HandleFunc("/api/chirps", apiCfg.getChirpsHandler).Methods("GET")
	router.HandleFunc("/api/chirps/{chirpID}", apiCfg.getUserChirpsHandler).Methods("GET")
	router.HandleFunc("/api/chirps/{chirpID}", apiCfg.deleteChirpHandler).Methods("DELETE")

	// Users
	router.HandleFunc("/api/users", apiCfg.createUserHandler).Methods("POST")
	router.HandleFunc("/api/users", apiCfg.updateUserHandler).Methods("PUT")

	// Authentication
	router.HandleFunc("/api/login", apiCfg.loginHandler).Methods("POST")
	router.HandleFunc("/api/refresh", apiCfg.refreshHandler).Methods("POST")
	router.HandleFunc("/api/revoke", apiCfg.revokeHandler).Methods("POST")

	// Admin routes
	router.HandleFunc("/admin/reset", apiCfg.adminResetHandler).Methods("POST")
	router.HandleFunc("/admin/metrics", apiCfg.handlerMetrics).Methods("GET")
	
	// Polka Webhooks
	router.HandleFunc("/api/polka/webhooks", apiCfg.polkaWebhookHandler).Methods("POST")

	// Start the server
	fmt.Println("Serving files on http://localhost:8080/app/")
	fmt.Println("Static files available at http://localhost:8080/app/")
	fmt.Println("Reset available at http://localhost:8080/admin/reset")
	fmt.Println("Users available at http://localhost:8080/api/users")
	fmt.Println("Update user available at http://localhost:8080/api/users")
	fmt.Println("All chirps available at http://localhost:8080/api/chirps")
	fmt.Println("User chirps available at http://localhost:8080/api/chirps/{chirpID}")
	fmt.Println("Health check available at http://localhost:8080/admin/healthz")
	fmt.Println("Metrics available at http://localhost:8080/admin/metrics")
	fmt.Println("Token refresh available at http://localhost:8080//api/refresh")
	fmt.Println("Token revoke available at http://localhost:8080//api/revoke")

	// Wrap mux with the logging middleware: (optional)
	muxWithLogging := loggingMiddleware(router)
	http.ListenAndServe(":8080", muxWithLogging)
	fmt.Println("Starting server on :8080")

}
