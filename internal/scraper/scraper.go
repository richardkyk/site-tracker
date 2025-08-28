package scraper

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"site-tracker/internal/clients/s3"
	"site-tracker/internal/models"

	"github.com/gocolly/colly/v2"
)

func Scrape(ctx context.Context, site models.Site) (string, error) {
	c := colly.NewCollector()

	var outputError error
	var extractedValue string
	var htmlContent string

	c.OnHTML(site.Selector, func(e *colly.HTMLElement) {
		text := e.Text
		valueRegex := regexp.MustCompile(site.Regex)
		match := valueRegex.FindStringSubmatch(text)
		extractedValue = "N/A"

		// Upload the HTML to S3
		_htmlContent, err := e.DOM.Html()
		if err != nil {
			log.Printf("failed to get HTML: %v", err)
			return
		}
		htmlContent = _htmlContent

		if len(match) > 1 {
			extractedValue = match[1]
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("request failed: %s %d %v", r.Request.URL, r.StatusCode, err)
		outputError = err
		if err != nil {
			log.Printf("failed to upload HTML: %v", err)
		}
	})

	c.Visit(site.URL)
	if outputError != nil {
		err := UploadHTML(ctx, site, htmlContent)
		if err != nil {
			log.Printf("failed to upload HTML: %v", err)
		}
		return "", outputError
	}

	if extractedValue == "" {
		err := UploadHTML(ctx, site, htmlContent)
		if err != nil {
			log.Printf("failed to upload HTML: %v", err)
		}
		return "", fmt.Errorf("selector not found")
	}
	log.Printf("extracted value: %s", extractedValue)

	return extractedValue, nil
}

func UploadHTML(ctx context.Context, site models.Site, htmlContent string) error {
	key := fmt.Sprintf("%s.html", site.ID)
	htmlBytes := []byte(htmlContent)
	if err := s3.UploadBytes(ctx, key, htmlBytes); err != nil {
		return err
	}
	log.Printf("uploaded %s to %s", key, site.ID)
	return nil
}
