package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

func RunServer() {
	const port = ":8080"
	r := gin.Default()

	r.GET("/movies", GetMovies)
	r.GET("/movies/:id", GetMovieByID)
	r.GET("/movie-images", GetMovieImages)

	r.Static("/static", "./web")

	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	r.GET("/movie/:id", ShowMoviePage)

	r.LoadHTMLGlob("web/template/*")

	log.Printf("Starting API server on %s\n", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

