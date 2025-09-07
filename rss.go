package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/heavysider/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var rssFeed RSSFeed
	err = xml.Unmarshal(respBody, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for i, rssFeedItem := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeedItem.Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeedItem.Description)
	}
	return &rssFeed, nil
}

func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("error determining which feed to fetch next: %v\n", err)
		return
	}
	fmt.Printf("Scraping feed: %v\n", feed.Name)

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("error fetching the feed %v: %v\n", feed.Name, err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("error marking feed %v as fetched\n", feed.Name)
		return
	}

	for _, rssItem := range rssFeed.Channel.Item {
		description := sql.NullString{
			String: rssItem.Description,
			Valid:  true,
		}
		parsedPublisedAt, err := parseTimeFlexible(rssItem.PubDate)
		if err != nil {
			continue
		}
		creationTime := time.Now()
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			Title:       rssItem.Title,
			Url:         rssItem.Link,
			Description: description,
			FeedID:      feed.ID,
			PublishedAt: parsedPublisedAt,
			CreatedAt:   creationTime,
			UpdatedAt:   creationTime,
		})
		if err != nil {
			if err.Error() != "pq: duplicate key value violates unique constraint \"posts_url_key\"" {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}
}

func parseTimeFlexible(timeStr string) (time.Time, error) {
	formats := []string{
		time.RFC3339, // "2006-01-02T15:04:05Z07:00"
		time.RFC822,  // "02 Jan 06 15:04 MST"
		time.RFC1123, // "Mon, 02 Jan 2006 15:04:05 MST"
		"2006-01-02 15:04:05",
		"2006-01-02",
		"01/02/2006",
		"01-02-2006",
		"2006/01/02",
		// Add more formats as needed
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}
