package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Ka10ken1/mykadri-scraper/internal/models"
)

func GetMovies(c *gin.Context) {
	movies, err := models.GetAllMovies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movies"})
		return
	}
	c.JSON(http.StatusOK, movies)
}



func GetMovieByID(c *gin.Context) {
	id := c.Param("id")

	movie, err := models.GetMovieByID(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie"})
		return
	}

	if movie == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	c.JSON(http.StatusOK, movie)
}


func GetMovieImages(c *gin.Context) {
    images, err := models.GetAllMovieImages()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get movie images"})
        return
    }
    c.JSON(http.StatusOK, images)
}



func ShowMoviePage(c *gin.Context) {
	id := c.Param("id")
	movie, err := models.GetMovieByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Movie not found")
		return
	}

	c.HTML(http.StatusOK, "movie.html", gin.H{
		"Title":        movie.Title,
		"EnglishTitle": movie.TitleEnglish,
		"VideoURL":     movie.VideoURL,
		"Image":        movie.Image,
		"Year":         movie.Year,
	})
}

func GetMoviesByTitle(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	movies, err := models.SearchMoviesByTitle(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

