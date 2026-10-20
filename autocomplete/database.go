
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	_ "github.com/microsoft/go-mssqldb"
)

// Movie represents one record from MovieDB.
type Movie struct {
	ID          int
	Title       string
	ReleaseYear int
	Popularity  float64
}

// ConnectDB initializes the SQL Server connection pool.
func ConnectDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "1433")
	dbName := getEnv("DB_NAME", "MovieDB")
	user := getEnv("DB_USER", "sa")
	password := os.Getenv("DB_PASSWORD")

	if password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	params := url.Values{}
	params.Set("database", dbName)
	params.Set("encrypt", "disable") // Local development only

	connURL := &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		RawQuery: params.Encode(),
	}

	db, err := sql.Open("sqlserver", connURL.String())
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Connection pool configuration
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Verify that SQL Server is reachable.
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to SQL Server: %w", err)
	}

	return db, nil
}

// SearchMovies retrieves top-K matching movies.
// Current implementation: case-insensitive title prefix
// if the SQL Server column collation is case-insensitive.
func SearchMovies(
	ctx context.Context,
	db *sql.DB,
	prefix string,
	limit int,
) ([]Movie, error) {

	query := `
		SELECT TOP (@p1)
			id,
			title,
			release_year,
			popularity
		FROM dbo.Movies
		WHERE title COLLATE Latin1_General_100_CI_AS
		      LIKE @p2
		ORDER BY popularity DESC, id ASC
	`

	rows, err := db.QueryContext(
		ctx,
		query,
		limit,
		escapeLike(prefix)+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search movies: %w", err)
	}
	defer rows.Close()

	movies := make([]Movie, 0, limit)

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseYear,
			&movie.Popularity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan movie: %w", err)
		}

		movies = append(movies, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate movies: %w", err)
	}

	return movies, nil
}

// Escape SQL LIKE wildcard characters in user input.
func escapeLike(s string) string {
	result := ""
	for _, r := range s {
		switch r {
		case '%':
			result += "[%]"
		case '_':
			result += "[_]"
		case '[':
			result += "[[]"
		default:
			result += string(r)
		}
	}
	return result
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
