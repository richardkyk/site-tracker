package scraper

import (
	"fmt"
	"log"
	"regexp"
	"site-tracker/internal/models"

	"github.com/gocolly/colly/v2"
)

func Scrape(site models.Site) (string, error) {
	c := colly.NewCollector()

	var outputError error
	var extractedValue string

	c.OnHTML(site.Selector, func(e *colly.HTMLElement) {
		text := e.Text
		valueRegex := regexp.MustCompile(site.Regex)
		match := valueRegex.FindStringSubmatch(text)
		extractedValue = "N/A"

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
	})

	c.Visit(site.URL)
	if outputError != nil {
		return "", outputError
	}

	if extractedValue == "" {
		return "", fmt.Errorf("selector not found")
	}
	log.Printf("extracted value: %s", extractedValue)

	return extractedValue, nil
}
