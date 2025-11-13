package main

import (
	"fmt"
	"log"

	"github.com/mmcdole/gofeed"
)

func main() {

	feeds := []string{
        "https://www.eltiempo.com/rss/eltiempo.xml",
        "https://www.larepublica.co/rss",    
        "https://feeds.bbci.co.uk/news/rss.xml",
        "https://rss.nytimes.com/services/xml/rss/nyt/HomePage.xml",
        "https://www.theguardian.com/world/rss",
	}

	parser := gofeed.NewParser()

	for _, url := range feeds {
		fmt.Println("===============================================")
		fmt.Println("FEED:", url)
		fmt.Println("===============================================")

		feed, err := parser.ParseURL(url)
		if err != nil {
			log.Println("Error al leer feed:", url, err)
			continue
		}

		fmt.Println("Título del canal:", feed.Title)
		fmt.Println("Descripción:", feed.Description)
		fmt.Println("Primeros 3 artículos:")

		limit := 3
		if len(feed.Items) < limit {
			limit = len(feed.Items)
		}

		for i := 0; i < limit; i++ {
			item := feed.Items[i]
			fmt.Println("---------------")

			fmt.Println("Title:", item.Title)
			fmt.Println("Link:", item.Link)
			fmt.Println("Published:", item.Published)
			fmt.Println("Description:", truncate(item.Description, 120))

			if len(item.Categories) > 0 {
				fmt.Println("Categories:", item.Categories)
			}

			if item.Author != nil {
				fmt.Println("Author:", item.Author.Name)
			}

			if item.Image != nil {
				fmt.Println("Image (via Image struct):", item.Image.URL)
			}

			if len(item.Enclosures) > 0 {
				fmt.Println("Enclosure:", item.Enclosures[0].URL)
			}
		}
	}
}

// Helper: corta la descripción para imprimir limpio
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

