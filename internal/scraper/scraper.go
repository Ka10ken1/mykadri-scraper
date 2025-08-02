package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/Ka10ken1/mykadri-scraper/internal/models"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type Movie = models.Movie

type Status int

const (
	Scraped Status = iota
	NeedsTobeScraped
)


func ScrapeMovies(client *http.Client) ([]Movie, error) {
	alreadyScraped, err := models.HasMovies()
	if err != nil {
		return nil, fmt.Errorf("db check error: %w", err)
	}
	if alreadyScraped {
		fmt.Println("Movies already scraped, skipping scraping.")
		return nil, nil
	}

	existingLinks, err := models.GetAllMovieLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to preload movie links: %w", err)
	}

	seen := make(map[string]struct{}, len(existingLinks))
	for _, link := range existingLinks {
		seen[link] = struct{}{}
	}

	c := setupCollector(client)

	var mu sync.Mutex
	var movies []Movie

	c.OnHTML("div.post.post-t1", func(e *colly.HTMLElement) {
		movie := parseMovie(e)
		log.Printf("Found movie: %s (%s)", movie.Title, movie.Year)

		videoURL, err := scrapeVideoURL(client, movie.Link)

		if err != nil || videoURL == "" {
			log.Printf("Warning: could not get video URL for %s: %v", movie.Title, err)
			return 
		}
		
		movie.VideoURL = videoURL

		mu.Lock()
		if _, found := seen[movie.Link]; !found {
			movies = append(movies, movie)
			seen[movie.Link] = struct{}{}
		}
		mu.Unlock()
	})


	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})


	c.OnError(func(r *colly.Response, err error) {
		if r != nil && r.StatusCode == 429 {
			log.Printf("Error 429 on %s, skipping...", r.Request.URL)
		} else if r != nil {
			log.Printf("Request error on %s: %v", r.Request.URL, err)
		} else {
			log.Printf("Request error: %v", err)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode == 429 {
			log.Printf("Received 429 on %s, skipping...", r.Request.URL)
		}
	})


	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "mykadri.tv",
		Parallelism: 2,
		Delay:       500 * time.Millisecond,
		RandomDelay: 200 * time.Microsecond,
	})

	if err != nil {
		return nil, err
	}

	

	baseURL := "https://mykadri.tv/filmebi_qartulad/page/%d/"
	maxPages := 332

	var wg sync.WaitGroup
	for i := 1; i <= maxPages; i++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			url := fmt.Sprintf(baseURL, page)
			if err := c.Visit(url); err != nil {
				log.Println("Failed to visit", url, err)
			}
		}(i)
	}
	wg.Wait()
	c.Wait()

	return movies, nil
}


func setupCollector(client *http.Client) *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains("mykadri.tv", "www.mykadri.tv"),
		colly.Async(true),
	)

	c.SetClient(client)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
	})

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	return c
}

func parseMovie(e *colly.HTMLElement) Movie {
	title := e.DOM.Find("a.post-link.post-title-primary").AttrOr("title", "")
	link := e.Request.AbsoluteURL(e.DOM.Find("a.post-link.post-title-primary").AttrOr("href", ""))

	year := ""
	secondaryTitle := e.DOM.Find("a.post-link.post-title-secondary").Text()
	re := regexp.MustCompile(`\((\d{4})\)`)
	match := re.FindStringSubmatch(secondaryTitle)
	if len(match) > 1 {
		year = match[1]
	} else {
		year = e.DOM.Find("div.yearshort > span.left").Text()
	}

	img := e.DOM.Find("div.post-image-wrapper img.post-image")
	imgURL, exists := img.Attr("data-lazy")
	if !exists {
		imgURL = img.AttrOr("src", "")
	}
	imgURL = e.Request.AbsoluteURL(imgURL)

	return Movie{
		Title: title,
		Year:  year,
		Link:  link,
		Image: imgURL,
	}
}

func scrapeVideoURL(client *http.Client, moviePageURL string) (string, error) {

	resp, err := client.Get(moviePageURL)
	if err != nil {
		return "", fmt.Errorf("failed to GET page: %w", err)
	}
	defer resp.Body.Close()


	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}
	body := string(bodyBytes)

	re := regexp.MustCompile(`data-lazy="(https://vidsrc\.me/embed/movie\?imdb=tt\d+)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("video URL not found in HTML")
	}
	url := matches[1]
	return url, nil

}

