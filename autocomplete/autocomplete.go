package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultLimit = 10
	maxLimit     = 20
	maxQueryLen  = 100
)

type MovieSuggestion struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	ReleaseYear int     `json:"release_year"`
	Popularity  float64 `json:"popularity"`
}

type Pagination struct {
	Limit    int `json:"limit"`
	Returned int `json:"returned"`
}

type AutocompleteResponse struct {
	Query           string            `json:"query"`
	NormalizedQuery string            `json:"normalized_query"`
	Suggestions     []MovieSuggestion `json:"suggestions"`
	Pagination      Pagination        `json:"pagination"`
}

type APIError struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func main() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to SQL Server: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/v1/autocomplete", autocompleteHandler(db))

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Autocomplete API listening on :8080")

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

// healthHandler checks whether the HTTP service is running.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed,
			"METHOD_NOT_ALLOWED", "Only GET is supported")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// autocompleteHandler handles movie autocomplete requests.
func autocompleteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed,
				"METHOD_NOT_ALLOWED", "Only GET is supported")
			return
		}

		// 1. Read and validate query
		query := r.URL.Query().Get("q")
		normalizedQuery := strings.ToLower(strings.TrimSpace(query))

		if normalizedQuery == "" ||
			len([]rune(normalizedQuery)) > maxQueryLen {
			writeError(w, http.StatusBadRequest,
				"INVALID_QUERY",
				"Query must contain between 1 and 100 characters")
			return
		}

		// 2. Parse limit
		limit := defaultLimit

		if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
			parsedLimit, err := strconv.Atoi(rawLimit)

			if err != nil || parsedLimit < 1 ||
				parsedLimit > maxLimit {
				writeError(w, http.StatusBadRequest,
					"INVALID_LIMIT",
					"Limit must be between 1 and 20")
				return
			}

			limit = parsedLimit
		}

		// 3. Set query timeout
		ctx, cancel := context.WithTimeout(
			r.Context(),
			3*time.Second,
		)
		defer cancel()

		// 4. Retrieve movies
		movies, err := SearchMovies(
			ctx,
			db,
			normalizedQuery,
			limit,
		)

		if err != nil {
			log.Printf("Movie search failed: %v", err)

			if errors.Is(err, context.DeadlineExceeded) {
				writeError(w, http.StatusServiceUnavailable,
					"QUERY_TIMEOUT", "Movie search timed out")
				return
			}

			writeError(w, http.StatusServiceUnavailable,
				"DATABASE_ERROR",
				"Movie search is temporarily unavailable")
			return
		}

		// 5. Map database models to API models
		suggestions := make([]MovieSuggestion, 0, len(movies))

		for _, movie := range movies {
			suggestions = append(suggestions, MovieSuggestion{
				ID:          "movie:" + strconv.Itoa(movie.ID),
				Title:       movie.Title,
				ReleaseYear: movie.ReleaseYear,
				Popularity:  movie.Popularity,
			})
		}

		// 6. Construct API response
		response := AutocompleteResponse{
			Query:           query,
			NormalizedQuery: normalizedQuery,
			Suggestions:     suggestions,
			Pagination: Pagination{
				Limit:    limit,
				Returned: len(suggestions),
			},
		}

		writeJSON(w, http.StatusOK, response)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON: %v", err)
	}
}

func writeError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	writeJSON(w, status, APIError{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
