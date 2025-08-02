package main

import (
    "log"
    "os"
    "github.com/Ka10ken1/mykadri-scraper/internal/models"
    "github.com/Ka10ken1/mykadri-scraper/internal/scraper"
    "github.com/Ka10ken1/mykadri-scraper/internal/tui"
    "github.com/joho/godotenv"
)


func main() {

    if err := godotenv.Load(); err != nil {
	log.Println("No .env file found, using default env vars")
    }

    uri := os.Getenv("MONGO_URI")
    db := os.Getenv("MONGO_DB")
    coll := os.Getenv("MONGO_COLLECTION")

    if err := models.InitMongo(uri, db, coll); err != nil {
	log.Fatal(err)
    }

    if err := models.EnsureTextIndex(); err != nil {
	log.Fatalf("Failed to create text index: %v", err)
    }

    movies, err := scraper.ScrapeMovies()
    if err != nil {
	log.Fatal(err)
    }

    if len(movies) > 0 {
	if err := models.InsertMovies(movies); err != nil {
	    log.Fatal(err)
	}
    } else {
	log.Println("No new movies to insert, skipping DB insert.")
    }

    tui.Run()
}

