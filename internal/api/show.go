package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Ka10ken1/mykadri-scraper/internal/models"
)

func GetShows(c *gin.Context) {
	shows, err := models.GetAllShows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get shows"})
		return
	}
	c.JSON(http.StatusOK, shows)
}

func GetShowByID(c *gin.Context) {
	id := c.Param("id")

	show, err := models.GetShowByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get show"})
		return
	}

	if show == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "show not found"})
		return
	}

	c.JSON(http.StatusOK, show)
}

func GetShowImages(c *gin.Context) {
	images, err := models.GetAllShowImages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get show images"})
		return
	}
	c.JSON(http.StatusOK, images)
}

func ShowShowPage(c *gin.Context) {
	id := c.Param("id")
	show, err := models.GetShowByID(id)
	if err != nil {
		c.String(http.StatusNotFound, "Show not found")
		return
	}

	c.HTML(http.StatusOK, "show.html", gin.H{
		"Title":        show.Title,
		"EnglishTitle": show.TitleEnglish,
		"VideoURL":     show.VideoURL,
		"Image":        show.Image,
		"Year":         show.Year,
	})
}

func GetShowsByTitle(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	shows, err := models.SearchShowsByTitle(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search shows"})
		return
	}

	c.JSON(http.StatusOK, shows)
}

