package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Ka10ken1/mykadri-scraper/internal/api"
	"github.com/Ka10ken1/mykadri-scraper/internal/models"
	"github.com/Ka10ken1/mykadri-scraper/internal/scraper"
	"github.com/joho/godotenv"
)

func createHTTPClientWithCustomDNS() *http.Client {
    resolver := &net.Resolver{
        PreferGo: true,
        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
            d := net.Dialer{
                Timeout: time.Second * 2,
            }
            return d.DialContext(ctx, network, "1.1.1.1:53")
        },
    }

    dialer := &net.Dialer{
        Resolver: resolver,
        Timeout:  5 * time.Second,
    }

    transport := &http.Transport{
        DialContext: dialer.DialContext,
    }

    return &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second,
    }
}


func main() {

    client := createHTTPClientWithCustomDNS()

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

    movies, err := scraper.ScrapeMovies(client)
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

    api.RunServer()


}

