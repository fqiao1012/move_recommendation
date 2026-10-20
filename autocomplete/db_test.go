
package main

import (
	"context"
	"testing"
	"time"
)

func TestSearchMovies(t *testing.T) {
	db, err := ConnectDB()
	if err != nil {
		t.Fatalf("DB connection failed: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	movies, err := SearchMovies(
		ctx,
		db,
		"star",
		10,
	)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(movies) == 0 {
		t.Fatal("Expected movies matching 'star'")
	}

	for _, movie := range movies {
		t.Logf(
			"Movie: %s, Popularity: %.2f",
			movie.Title,
			movie.Popularity,
		)
	}
}
