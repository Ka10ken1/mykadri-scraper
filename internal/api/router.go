package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

func RunServer() {
	const port = ":8080"
	r := gin.Default()

	r.GET("/api/movies", GetMovies)
	r.GET("/api/movies/:id", GetMovieByID)
	r.GET("/api/movie-images", GetMovieImages)
	r.GET("/api/search", GetMoviesByTitle)
	r.GET("/api/movie/:id", ShowMoviePage)

	r.GET("/api/shows", GetShows)
	r.GET("/api/shows/:id", GetShowByID)
	r.GET("/api/shows/images", GetShowImages)
	r.GET("/api/shows/search", GetShowsByTitle)
	r.GET("/api/show/:id", ShowShowPage)

	r.Static("/static", "./web")

	
	r.GET("/", func(c *gin.Context) {
		c.File("./web/media.html")
	})


	r.GET("/movies", func(c *gin.Context) {
		c.File("./web/movies.html")
	})

	r.GET("/shows", func(c *gin.Context) {
		c.File("./web/shows.html")
	})

	r.LoadHTMLGlob("web/template/*")

	log.Printf("Starting API server on %s\n", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

