package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmacneill66/go_projects/chirpy/internal/auth"
	"github.com/jmacneill66/go_projects/chirpy/internal/database"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

/*
// Middleware to increment the request count (optional)
// This is a simple example of how you might implement a middleware to count requests.

	func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg.fileserverHits.Add(1) // ✅ Safely increment the counter
			next.ServeHTTP(w, r)
		})
	}
*/
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load() // Get visit count
	// Dynamically inject 'hits' into the HTML template
	html := fmt.Sprintf(`
	<html>
	  <body>
	    <h1>Welcome, Chirpy Admin</h1>
	    <p>Chirpy has been visited %d times!</p>
	  </body>
	</html>
	`, hits)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// Health check handler (allows all HTTP methods)
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) adminResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}
	err := cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to reset database"})
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Create user handler
func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}
	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}
	// Store user in DB
	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
		return
	}
	// ✅ Manually map to a response struct with proper JSON tags (to pass tests, expecting camelCase)
	response := struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: false, // Default value or replace with a valid field if added to CreateUserRow
	}
	// Respond (excluding password)
	RespondWithJSON(w, http.StatusCreated, response)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		//ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}
	// Get user by email
	user, err := cfg.DB.GetUserByEmail(r.Context(), req.Email)
	if err != nil || auth.CheckPasswordHash(user.HashedPassword, req.Password) != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Incorrect email or password"})
		return
	}
	// Generate JWT (expires in 1 hour)
	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating token"})
		return
	}
	// Generate Refresh Token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating token"})
		return
	}
	// Store Refresh Token (expires in 60 days)
	expiresAt := time.Now().Add(60 * 24 * time.Hour)
	err = cfg.DB.StoreRefreshToken(r.Context(), database.StoreRefreshTokenParams{
		Token:     refreshToken,
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		ExpiresAt: expiresAt,
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error creating token"})
		return
	}
	// Response
	response := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    string    `json:"created_at"`
		UpdatedAt    string    `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    user.UpdatedAt.Format(time.RFC3339),
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed.Valid && user.IsChirpyRed.Bool,
	}
	json.NewEncoder(w).Encode(response)
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from headers
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	// Get user from refresh token
	userID, err := getUserFromRefreshToken(r.Context(), tokenStr, cfg)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	// Generate new JWT
	newToken, err := auth.MakeJWT(userID, cfg.JWTSecret, time.Hour)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	// Response
	response := struct {
		Token string `json:"token"`
	}{
		Token: newToken,
	}
	json.NewEncoder(w).Encode(response)
}

func getUserFromRefreshToken(ctx context.Context, tokenStr string, apiCfg *apiConfig) (uuid.UUID, error) {
	// Example implementation: Retrieve user ID from the database using the refresh token
	userID, err := apiCfg.DB.GetUserIDByRefreshToken(ctx, tokenStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid refresh token")
	}
	if !userID.Valid {
		return uuid.Nil, fmt.Errorf("invalid refresh token")
	}
	return userID.UUID, nil
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from headers
	tokenStr, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		return
	}
	// Revoke token in DB
	err = cfg.DB.RevokeRefreshToken(r.Context(), tokenStr)
	if err != nil {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Failed to revoke token"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Update user handler
func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
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
	// Parse request body
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		return
	}
	// Hash the new password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
		return
	}
	// Update user in the database
	err = cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update user"})
		return
	}
	// Fetch updated user details
	user, err := cfg.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch updated user"})
		return
	}
	// Respond with updated user details (excluding password)
	response := struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   string    `json:"created_at"`
		UpdatedAt   string    `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Valid && user.IsChirpyRed.Bool,
	}
	RespondWithJSON(w, http.StatusOK, response)
}
func (cfg *apiConfig) polkaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Step 1: Validate the API key
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.PolkaKey {
		RespondWithJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid API key"})
		return
	}
	// Step 2: Parse the request body
	var webhookRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}
	err = json.NewDecoder(r.Body).Decode(&webhookRequest)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid JSON"})
		return
	}
	// Step 3: Ignore events that are not "user.upgraded"
	if webhookRequest.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Step 4: Convert User ID to UUID and upgrade the user
	userID, err := uuid.Parse(webhookRequest.Data.UserID)
	if err != nil {
		RespondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid user ID format"})
		return
	}
	rowsAffected, err := cfg.DB.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": "Database error"})
		return
	}
	// Step 5: If no rows were affected, return 404 (User not found)
	if rowsAffected == 0 {
		RespondWithJSON(w, http.StatusNotFound, map[string]string{"error": "User not found"})
		return
	}
	// Step 6: Return 204 No Content on success
	w.WriteHeader(http.StatusNoContent)
}
