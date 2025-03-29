package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmacneill66/go_projects/chirpy/internal/auth"
	"github.com/jmacneill66/go_projects/chirpy/internal/database"
)

// List of words to censor
var badWords = []string{"kerfuffle", "sharbert", "fornax"}

// Struct to represent the JSON request body
type ChirpRequest struct {
	Body string `json:"body"`
}

// Struct for error responses
type ErrorResponse struct {
	Error string `json:"error"`
}

// Struct for valid response
type ValidResponse struct {
	Valid bool `json:"valid"`
}

// Struct for response with cleaned chirp
type CleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

// Struct for chirp response
type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func dbChirpToResponse(dbChirp database.Chirp) ChirpResponse {
	return ChirpResponse{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: dbChirp.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

// Helper: Respond with JSON
func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Helper: Censor bad words
func censorBadWords(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		// Convert to lowercase for case-insensitive match
		loweredWord := strings.ToLower(word)
		for _, badWord := range badWords {
			if loweredWord == badWord {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}

// Handler to create and validate Chirp
func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a POST request
	if r.Method != http.MethodPost {
		RespondWithJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method Not Allowed"})
		return
	}
	// Parse JSON request
	var requestBody struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	// Validate Chirp length (≤ 140 characters)
	if len(requestBody.Body) > 140 {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Chirp is too long"})
		return
	}
	// Censor Profane Words
	cleanedBody := censorBadWords(requestBody.Body)
	// Get JWT from request headers
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Validate JWT and extract user ID
	userID, err := auth.ValidateJWT(tokenStr, cfg.JWTSecret)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	// Insert Chirp into Database
	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create chirp"})
		return
	}
	response := dbChirpToResponse(chirp)
	RespondWithJSON(w, http.StatusCreated, response)
}

// Handler to get Chirps
func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a GET request
	if r.Method != http.MethodGet {
		RespondWithJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Not a GET request"})
		return
	}
	// Extract optional query parameters
	authorIDStr := r.URL.Query().Get("author_id")
	sortOrder := r.URL.Query().Get("sort") // "asc" or "desc" (default: "asc")

	var chirps []database.Chirp
	var err error

	if authorIDStr != "" {
		// Parse author_id as UUID
		authorID, err := uuid.Parse(authorIDStr)
		if err != nil {
			RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid author_id"})
			return
		}
		// Fetch chirps for a specific author
		chirpRows, err := cfg.DB.GetChirpsByAuthor(r.Context(), authorID)
		if err != nil {
			RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get chirps"})
			return
		}
		// Convert []database.GetChirpsByAuthorRow to []database.Chirp
		for _, row := range chirpRows {
			chirps = append(chirps, database.Chirp{
				ID:        row.ID,
				UserID:    row.UserID,
				Body:      row.Body,
				CreatedAt: row.CreatedAt,
			})
		}
	} else {
		// Fetch all chirps
		chirps, err = cfg.DB.GetAllChirps(r.Context())
		if err != nil {
			RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get chirps"})
			return
		}
	}
	// Sort chirps by created_at
	if sortOrder == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
		})
	}
	// Convert database chirps to response format
	var response []ChirpResponse
	for _, chirp := range chirps {
		response = append(response, dbChirpToResponse(chirp))
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// Handler to get a single chirp by ID
func (cfg *apiConfig) getUserChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract chirpID from the path parameter
	chirpIDStr := mux.Vars(r)["chirpID"] // chirpIDStr := r.PathValue("chirpID")
	// Parse chirpID as UUID
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid chirp ID"})
		return
	}
	// Fetch chirp from the database
	chirp, err := cfg.DB.GetChirpsForUsers(r.Context(), chirpID)
	if err != nil {
		RespondWithJSON(w, http.StatusNotFound, map[string]string{"error": "Chirp not found"})
		return
	}
	// Convert database chirp to response format
	response := dbChirpToResponse(chirp)
	// Return the chirp
	RespondWithJSON(w, http.StatusOK, response)
}

// deletere chirp by ID
func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	// Extract and verify the access token
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	// Extract user ID from the token
	userID, err := auth.ValidateJWT(tokenStr, cfg.JWTSecret)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		return
	}
	// Get chirpID from URL variables
	vars := mux.Vars(r)
	chirpIDStr := vars["chirpID"]
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid chirp ID"})
		return
	}
	// Check if the chirp exists and is owned by the user
	chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		RespondWithJSON(w, http.StatusNotFound, map[string]string{"error": "Chirp not found"})
		return
	}
	// Ensure the requesting user is the author of the chirp
	if chirp.UserID != userID {
		RespondWithJSON(w, http.StatusForbidden, map[string]string{"error": "You can only delete your own chirps"})
		return
	}
	// Delete the chirp
	err = cfg.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpID,
		UserID: userID,
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete chirp"})
		return
	}
	// Return 204 No Content (successful deletion)
	w.WriteHeader(http.StatusNoContent)
}
