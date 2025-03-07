// pkg/content/rss/parser.go
package rss

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/your-username/podcast-platform/pkg/content/models"
)

// Parser defines the interface for RSS feed parser
type Parser interface {
	ParseFeed(ctx context.Context, url string) (*models.RSSFeed, error)
}

type parser struct {
	httpClient *http.Client
}

// NewParser creates a new RSS feed parser
func NewParser(timeout time.Duration) Parser {
	return &parser{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// RSS feed structures
type rssChannel struct {
	XMLName     xml.Name    `xml:"channel"`
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Link        string      `xml:"link"`
	Language    string      `xml:"language"`
	Copyright   string      `xml:"copyright"`
	PubDate     string      `xml:"pubDate"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Category    []string    `xml:"category"`
	Generator   string      `xml:"generator"`
	Image       rssImage    `xml:"image"`
	Author      string      `xml:"author"`
	Owner       rssOwner    `xml:"itunes:owner"`
	Categories  []rssCategory `xml:"itunes:category"`
	Explicit    string      `xml:"itunes:explicit"`
	Items       []rssItem   `xml:"item"`
	ItunesImage itunesImage `xml:"itunes:image"`
	ItunesAuthor string   `xml:"itunes:author"`
	ItunesSummary string   `xml:"itunes:summary"`
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssImage struct {
	URL   string `xml:"url"`
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type itunesImage struct {
	Href string `xml:"href,attr"`
}

type rssOwner struct {
	Name  string `xml:"itunes:name"`
	Email string `xml:"itunes:email"`
}

type rssCategory struct {
	Text     string        `xml:",chardata"`
	AttrText string        `xml:"text,attr"`
	Category *rssSubcategory `xml:"itunes:category"`
}

type rssSubcategory struct {
	Text string `xml:"text,attr"`
}

type rssEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

type rssItem struct {
	Title           string        `xml:"title"`
	Description     string        `xml:"description"`
	Link            string        `xml:"link"`
	Guid            string        `xml:"guid"`
	PubDate         string        `xml:"pubDate"`
	Duration        string        `xml:"itunes:duration"`
	Author          string        `xml:"itunes:author"`
	Subtitle        string        `xml:"itunes:subtitle"`
	Summary         string        `xml:"itunes:summary"`
	Enclosure       rssEnclosure  `xml:"enclosure"`
	ItunesImage     itunesImage   `xml:"itunes:image"`
	ItunesEpisode   string        `xml:"itunes:episode"`
	ItunesSeason    string        `xml:"itunes:season"`
	Content         string        `xml:"content:encoded"`
	Explicit        string        `xml:"itunes:explicit"`
}

// ParseFeed parses an RSS feed from a URL
func (p *parser) ParseFeed(ctx context.Context, url string) (*models.RSSFeed, error) {
	// Create a request with the provided context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set appropriate headers
	req.Header.Set("User-Agent", "Sudanese Podcast Platform RSS Parser/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	// Make the request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed request failed with status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read feed body: %w", err)
	}

	// Parse the XML
	var feed rssFeed
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.Strict = false // Be lenient with XML parsing errors
	if err := decoder.Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to parse feed XML: %w", err)
	}

	// Check if feed has content
	if feed.Channel.Title == "" {
		return nil, errors.New("feed has no content or is not a valid podcast feed")
	}

	// Convert RSS feed to our model
	result := &models.RSSFeed{
		Title:        feed.Channel.Title,
		Description:  feed.Channel.Description,
		Language:     feed.Channel.Language,
		WebsiteURL:   feed.Channel.Link,
		Explicit:     parseBooleanString(feed.Channel.Explicit),
	}

	// Get main category and subcategory
	if len(feed.Channel.Categories) > 0 {
		mainCategory := feed.Channel.Categories[0]
		
		// Try to get the category from different attribute locations
		if mainCategory.AttrText != "" {
			result.Category = mainCategory.AttrText
		} else if mainCategory.Text != "" {
			result.Category = mainCategory.Text
		}
		
		// Check for subcategory
		if mainCategory.Category != nil && mainCategory.Category.Text != "" {
			result.Subcategory = mainCategory.Category.Text
		}
	}

	// Set the author from various possible fields
	if feed.Channel.Author != "" {
		result.Author = feed.Channel.Author
	} else if feed.Channel.ItunesAuthor != "" {
		result.Author = feed.Channel.ItunesAuthor
	} else if feed.Channel.Owner.Name != "" {
		result.Author = feed.Channel.Owner.Name
	} else {
		result.Author = feed.Channel.Title // Fallback to title
	}

	// Get cover image URL
	if feed.Channel.ItunesImage.Href != "" {
		result.CoverImageURL = feed.Channel.ItunesImage.Href
	} else if feed.Channel.Image.URL != "" {
		result.CoverImageURL = feed.Channel.Image.URL
	}

	// Parse episodes
	result.Items = make([]models.RSSFeedItem, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		// Skip items without enclosures or GUIDs
		if item.Enclosure.URL == "" || item.Guid == "" {
			continue
		}
		
		// Parse episode
		episode := models.RSSFeedItem{
			Title:       item.Title,
			GUID:        item.Guid,
			AudioURL:    item.Enclosure.URL,
		}
		
		// Get description from various possible fields
		if item.Summary != "" {
			episode.Description = item.Summary
		} else if item.Description != "" {
			episode.Description = item.Description
		} else if item.Content != "" {
			episode.Description = item.Content
		}
		
		// Clean up HTML tags in description
		episode.Description = cleanHTMLContent(episode.Description)
		
		// Parse duration
		episode.Duration = parseDuration(item.Duration)
		
		// Parse publication date
		pubDate, err := parsePubDate(item.PubDate)
		if err == nil {
			episode.PublicationDate = pubDate
		} else {
			episode.PublicationDate = time.Now() // Fallback to current time
		}
		
		// Get episode cover image
		if item.ItunesImage.Href != "" {
			episode.CoverImageURL = item.ItunesImage.Href
		} else {
			episode.CoverImageURL = result.CoverImageURL // Fallback to podcast image
		}
		
		// Parse episode and season numbers
		if item.ItunesEpisode != "" {
			episodeNum, err := strconv.Atoi(item.ItunesEpisode)
			if err == nil {
				episode.EpisodeNumber = &episodeNum
			}
		}
		
		if item.ItunesSeason != "" {
			seasonNum, err := strconv.Atoi(item.ItunesSeason)
			if err == nil {
				episode.SeasonNumber = &seasonNum
			}
		}
		
		result.Items = append(result.Items, episode)
	}

	return result, nil
}

// parseDuration parses a duration string in various formats
// (e.g. "HH:MM:SS", "MM:SS", or seconds) to seconds
func parseDuration(duration string) int {
	if duration == "" {
		return 0
	}

	// Check if it's a plain number of seconds
	seconds, err := strconv.Atoi(duration)
	if err == nil {
		return seconds
	}

	// Try parsing "HH:MM:SS" or "MM:SS" format
	parts := strings.Split(duration, ":")
	var total int

	if len(parts) == 3 {
		// HH:MM:SS
		hours, _ := strconv.Atoi(parts[0])
		minutes, _ := strconv.Atoi(parts[1])
		seconds, _ := strconv.Atoi(parts[2])
		total = hours*3600 + minutes*60 + seconds
	} else if len(parts) == 2 {
		// MM:SS
		minutes, _ := strconv.Atoi(parts[0])
		seconds, _ := strconv.Atoi(parts[1])
		total = minutes*60 + seconds
	}

	return total
}

// parsePubDate parses publication date in various RFC formats
func parsePubDate(pubDate string) (time.Time, error) {
	if pubDate == "" {
		return time.Time{}, errors.New("empty publication date")
	}

	// Try different time formats
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		t, err := time.Parse(format, pubDate)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", pubDate)
}

// parseBooleanString parses itunes:explicit and similar boolean strings
func parseBooleanString(s string) bool {
	s = strings.ToLower(s)
	return s == "yes" || s == "true" || s == "1"
}

// cleanHTMLContent removes HTML tags from a string
func cleanHTMLContent(content string) string {
	// Replace HTML line breaks with newlines
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")
	
	// Remove HTML tags
	re := regexp.MustCompile("<[^>]*>")
	content = re.ReplaceAllString(content, "")
	
	// Decode HTML entities
	content = decodeHTMLEntities(content)
	
	// Trim whitespace
	return strings.TrimSpace(content)
}

// decodeHTMLEntities decodes common HTML entities
func decodeHTMLEntities(content string) string {
	entities := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
		"&nbsp;": " ",
	}
	
	for entity, replacement := range entities {
		content = strings.ReplaceAll(content, entity, replacement)
	}
	
	return content
}