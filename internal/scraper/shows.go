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
)


type Show = models.Show

func ScrapeShows(client *http.Client) ([]Show, error) {
	alreadyScraped, err := models.HasShows()
	if err != nil {
		return nil, fmt.Errorf("db check error: %w", err)
	}
	if alreadyScraped {
		log.Println("Shows already scraped, skipping scraping.")
		return nil, nil
	}

	existingLinks, err := models.GetAllShowLinks()
	if err != nil {
		return nil, fmt.Errorf("failed to preload show links: %w", err)
	}

	seen := make(map[string]struct{}, len(existingLinks))
	for _, link := range existingLinks {
		seen[link] = struct{}{}
	}

	c := setupCollector(client)

	var mu sync.Mutex
	var shows []Show

	c.OnHTML("div.post.post-t1", func(e *colly.HTMLElement) {
		show := parseShow(e)
		log.Printf("Found show: %s (%s)", show.Title, show.Year)

		videoURL, err := scrapeShowVideoURL(client, show.Link)
		if err != nil || videoURL == "" {
			log.Printf("Warning: could not get video URL for %s: %v", show.Title, err)
			return
		}

		show.VideoURL = videoURL

		mu.Lock()
		if _, found := seen[show.Link]; !found {
			shows = append(shows, show)
			seen[show.Link] = struct{}{}
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

	err = c.Limit(&colly.LimitRule{
		DomainGlob:  "mykadri.tv",
		Parallelism: 1,
		Delay:       2 * time.Second,
		RandomDelay: 500 * time.Microsecond,
	})
	if err != nil {
		return nil, err
	}

	baseURL := "https://mykadri.tv/serialebi_qartulad/page/%d/"
	maxPages := 38

	var wg sync.WaitGroup
	sema := make(chan struct{}, 1)

	for i := 1; i <= maxPages; i++ {
		wg.Add(1)
		sema <- struct{}{}

		go func(page int) {
			defer func() {
				<-sema
				wg.Done()
			}()
			url := fmt.Sprintf(baseURL, page)
			if err := c.Visit(url); err != nil {
				log.Println("Failed to visit", url, err)
			}
		}(i)
	}

	wg.Wait()
	c.Wait()

	return shows, nil
}

func parseShow(e *colly.HTMLElement) Show {
	title := e.DOM.Find("a.post-link.post-title-primary").AttrOr("title", "")
	englishTitle := e.DOM.Find("a.post-link.post-title-secondary").AttrOr("title", "")
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

	return Show{
		Title:        title,
		TitleEnglish: englishTitle,
		Year:         year,
		Link:         link,
		Image:        imgURL,
	}
}

func scrapeShowVideoURL(client *http.Client, showPageURL string) (string, error) {
	resp, err := client.Get(showPageURL)
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

	re := regexp.MustCompile(`data-lazy="(https://vidsrc\.me/embed/tv\?imdb=tt\d+)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("video URL not found in HTML")
	}

	return matches[1], nil
}

