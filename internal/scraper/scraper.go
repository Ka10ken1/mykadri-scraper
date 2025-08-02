package scraper

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"
	"github.com/Ka10ken1/mykadri-scraper/internal/models"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type Movie = models.Movie

type Status int

const (
	Scraped Status = iota
	NeedsTobeScraped
)


func ScrapeMovies() ([]Movie, error) {
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

	c := setupCollector()

	var mu sync.Mutex
	var movies []Movie

	c.OnHTML("div.post.post-t1", func(e *colly.HTMLElement) {
		movie := parseMovie(e)
		log.Printf("Found movie: %s (%s)", movie.Title, movie.Year)
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
		log.Printf("Request to %s failed with status %d: %v", r.Request.URL, r.StatusCode, err)
	})


	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "mykadri.tv",
		Parallelism: 10,
		Delay:       500 * time.Millisecond,
		RandomDelay: 200 * time.Microsecond,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set rate limit: %w", err)
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


func setupCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.AllowedDomains("mykadri.tv", "www.mykadri.tv"),
		colly.Async(true),
	)

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{}
					return d.DialContext(ctx, "tcp4", "8.8.8.8:53")
				},
			},
		}).DialContext,
	}

	c.WithTransport(transport)
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

