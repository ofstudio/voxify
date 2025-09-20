package feedcast

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestNewFeed(t *testing.T) {
	channelData := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories: []Category{
			NewCategory("Technology"),
		},
	}

	feed := NewFeed(channelData)

	if feed == nil {
		t.Fatal("NewFeed returned nil")
	}

	// Test basic structure
	if feed.xmlDoc.Version != xmlRssVersion {
		t.Errorf("Expected version %s, got %s", xmlRssVersion, feed.xmlDoc.Version)
	}

	if feed.xmlDoc.ItunesNS != xmlItunesNS {
		t.Errorf("Expected iTunes namespace %s, got %s", xmlItunesNS, feed.xmlDoc.ItunesNS)
	}

	if feed.xmlDoc.ContentNS != xmlContentNS {
		t.Errorf("Expected content namespace %s, got %s", xmlContentNS, feed.xmlDoc.ContentNS)
	}

	if feed.xmlDoc.PodcastNS != xmlPodcastNS {
		t.Errorf("Expected podcast namespace %s, got %s", xmlPodcastNS, feed.xmlDoc.PodcastNS)
	}

	// Test channel data
	channel := feed.xmlDoc.Channel
	if channel.Title != channelData.Title {
		t.Errorf("Expected title %s, got %s", channelData.Title, channel.Title)
	}

	if channel.Description.Data != channelData.Description {
		t.Errorf("Expected description %s, got %s", channelData.Description, channel.Description.Data)
	}

	if channel.ItunesImage.Href != channelData.Image {
		t.Errorf("Expected image %s, got %s", channelData.Image, channel.ItunesImage.Href)
	}

	if channel.Language != channelData.Language {
		t.Errorf("Expected language %s, got %s", channelData.Language, channel.Language)
	}

	if channel.ItunesExplicit != channelData.Explicit {
		t.Errorf("Expected explicit %s, got %s", channelData.Explicit, channel.ItunesExplicit)
	}

	if len(channel.ItunesCategory) != len(channelData.Categories) {
		t.Errorf("Expected %d categories, got %d", len(channelData.Categories), len(channel.ItunesCategory))
	}
}

func TestFeedValidation_RequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		channelData FeedData
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid feed",
			channelData: FeedData{
				Title:       "Valid Podcast",
				Description: "Valid description",
				Image:       "https://example.com/artwork.jpg",
				Language:    "en",
				Explicit:    ExplicitFalse,
				Categories:  []Category{NewCategory("Technology")},
			},
			expectError: true, // Because no items
			errorMsg:    "at least one channel item is required",
		},
		{
			name: "missing title",
			channelData: FeedData{
				Title:       "",
				Description: "Valid description",
				Image:       "https://example.com/artwork.jpg",
				Language:    "en",
				Explicit:    ExplicitFalse,
				Categories:  []Category{NewCategory("Technology")},
			},
			expectError: true,
			errorMsg:    "channel title is required",
		},
		{
			name: "missing description",
			channelData: FeedData{
				Title:       "Valid Podcast",
				Description: "",
				Image:       "https://example.com/artwork.jpg",
				Language:    "en",
				Explicit:    ExplicitFalse,
				Categories:  []Category{NewCategory("Technology")},
			},
			expectError: true,
			errorMsg:    "channel description is required",
		},
		{
			name: "missing image",
			channelData: FeedData{
				Title:       "Valid Podcast",
				Description: "Valid description",
				Image:       "",
				Language:    "en",
				Explicit:    ExplicitFalse,
				Categories:  []Category{NewCategory("Technology")},
			},
			expectError: true,
			errorMsg:    "channel itunes:image is required",
		},
		{
			name: "missing language",
			channelData: FeedData{
				Title:       "Valid Podcast",
				Description: "Valid description",
				Image:       "https://example.com/artwork.jpg",
				Language:    "",
				Explicit:    ExplicitFalse,
				Categories:  []Category{NewCategory("Technology")},
			},
			expectError: true,
			errorMsg:    "channel language is required",
		},
		{
			name: "invalid explicit value",
			channelData: FeedData{
				Title:       "Valid Podcast",
				Description: "Valid description",
				Image:       "https://example.com/artwork.jpg",
				Language:    "en",
				Explicit:    "invalid",
				Categories:  []Category{NewCategory("Technology")},
			},
			expectError: true,
			errorMsg:    "channel itunes:explicit must be either 'true' or 'false'",
		},
		{
			name: "missing categories",
			channelData: FeedData{
				Title:       "Valid Podcast",
				Description: "Valid description",
				Image:       "https://example.com/artwork.jpg",
				Language:    "en",
				Explicit:    ExplicitFalse,
				Categories:  []Category{},
			},
			expectError: true,
			errorMsg:    "at least one itunes:category is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feed := NewFeed(tt.channelData)
			err := feed.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFeedWithOptionalFields(t *testing.T) {
	channelData := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories:  []Category{NewCategory("Technology")},
	}

	feed := NewFeed(channelData)

	// Test WithAuthor
	author := "Test Author"
	feed.WithAuthor(author)
	if feed.xmlDoc.Channel.ItunesAuthor != author {
		t.Errorf("Expected author %s, got %s", author, feed.xmlDoc.Channel.ItunesAuthor)
	}

	// Test WithLink
	link := "https://example.com"
	feed.WithLink(link)
	if feed.xmlDoc.Channel.Link != link {
		t.Errorf("Expected link %s, got %s", link, feed.xmlDoc.Channel.Link)
	}

	// Test WithPubDate
	pubDate := time.Date(2023, 4, 1, 19, 0, 0, 0, time.UTC)
	feed.WithPubDate(pubDate)
	expectedPubDate := pubDate.Format(time.RFC1123Z)
	if feed.xmlDoc.Channel.PubDate != expectedPubDate {
		t.Errorf("Expected pubDate %s, got %s", expectedPubDate, feed.xmlDoc.Channel.PubDate)
	}

	// Test WithLastBuildDate
	lastBuildDate := time.Date(2023, 4, 2, 10, 0, 0, 0, time.UTC)
	feed.WithLastBuildDate(lastBuildDate)
	expectedLastBuildDate := lastBuildDate.Format(time.RFC1123Z)
	if feed.xmlDoc.Channel.LastBuildDate != expectedLastBuildDate {
		t.Errorf("Expected lastBuildDate %s, got %s", expectedLastBuildDate, feed.xmlDoc.Channel.LastBuildDate)
	}

	// Test WithItunesTitle
	itunesTitle := "iTunes Title"
	feed.WithItunesTitle(itunesTitle)
	if feed.xmlDoc.Channel.ItunesTitle != itunesTitle {
		t.Errorf("Expected iTunes title %s, got %s", itunesTitle, feed.xmlDoc.Channel.ItunesTitle)
	}

	// Test WithItunesType
	feed.WithItunesType(TypeSerial)
	if feed.xmlDoc.Channel.ItunesType != TypeSerial {
		t.Errorf("Expected iTunes type %s, got %s", TypeSerial, feed.xmlDoc.Channel.ItunesType)
	}

	// Test WithCopyright
	copyright := "© 2023 Test Company"
	feed.WithCopyright(copyright)
	if feed.xmlDoc.Channel.Copyright != copyright {
		t.Errorf("Expected copyright %s, got %s", copyright, feed.xmlDoc.Channel.Copyright)
	}

	// Test WithItunesNewFeedURL
	newFeedURL := "https://new-example.com/feed.xml"
	feed.WithItunesNewFeedURL(newFeedURL)
	if feed.xmlDoc.Channel.ItunesNewFeedURL != newFeedURL {
		t.Errorf("Expected new feed URL %s, got %s", newFeedURL, feed.xmlDoc.Channel.ItunesNewFeedURL)
	}

	// Test WithItunesBlock
	feed.WithItunesBlock(BlockYes)
	if feed.xmlDoc.Channel.ItunesBlock != BlockYes {
		t.Errorf("Expected iTunes block %s, got %s", BlockYes, feed.xmlDoc.Channel.ItunesBlock)
	}

	// Test WithItunesComplete
	feed.WithItunesComplete(CompleteYes)
	if feed.xmlDoc.Channel.ItunesComplete != CompleteYes {
		t.Errorf("Expected iTunes complete %s, got %s", CompleteYes, feed.xmlDoc.Channel.ItunesComplete)
	}

	// Test WithGenerator
	generator := "Voxify RSS Generator"
	feed.WithGenerator(generator)
	if feed.xmlDoc.Channel.Generator != generator {
		t.Errorf("Expected generator %s, got %s", generator, feed.xmlDoc.Channel.Generator)
	}

	// Test WithItunesSummary
	summary := "This is a podcast summary"
	feed.WithItunesSummary(summary)
	if feed.xmlDoc.Channel.ItunesSummary == nil || feed.xmlDoc.Channel.ItunesSummary.Data != summary {
		t.Errorf("Expected iTunes summary %s, got %v", summary, feed.xmlDoc.Channel.ItunesSummary)
	}

	// Test WithItunesKeywords
	keywords := "technology,podcast,apple"
	feed.WithItunesKeywords(keywords)
	if feed.xmlDoc.Channel.ItunesKeywords != keywords {
		t.Errorf("Expected iTunes keywords %s, got %s", keywords, feed.xmlDoc.Channel.ItunesKeywords)
	}

	// Test WithItunesOwner
	ownerName := "John Doe"
	ownerEmail := "john@example.com"
	feed.WithItunesOwner(ownerName, ownerEmail)
	if feed.xmlDoc.Channel.ItunesOwner == nil {
		t.Error("Expected iTunes owner to be set")
	} else {
		if feed.xmlDoc.Channel.ItunesOwner.Name != ownerName {
			t.Errorf("Expected owner name %s, got %s", ownerName, feed.xmlDoc.Channel.ItunesOwner.Name)
		}
		if feed.xmlDoc.Channel.ItunesOwner.Email != ownerEmail {
			t.Errorf("Expected owner email %s, got %s", ownerEmail, feed.xmlDoc.Channel.ItunesOwner.Email)
		}
	}
}

func TestFeedCategories(t *testing.T) {
	// Test single category
	channelData := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories:  []Category{NewCategory("Technology")},
	}

	feed := NewFeed(channelData)
	if len(feed.xmlDoc.Channel.ItunesCategory) != 1 {
		t.Errorf("Expected 1 category, got %d", len(feed.xmlDoc.Channel.ItunesCategory))
	}

	if feed.xmlDoc.Channel.ItunesCategory[0].Text != "Technology" {
		t.Errorf("Expected category 'Technology', got '%s'", feed.xmlDoc.Channel.ItunesCategory[0].Text)
	}

	// Test category with subcategories
	channelDataWithSub := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories: []Category{
			NewCategory("Society & Culture", "Documentary"),
		},
	}

	feedWithSub := NewFeed(channelDataWithSub)
	if len(feedWithSub.xmlDoc.Channel.ItunesCategory) != 1 {
		t.Errorf("Expected 1 category, got %d", len(feedWithSub.xmlDoc.Channel.ItunesCategory))
	}

	cat := feedWithSub.xmlDoc.Channel.ItunesCategory[0]
	if cat.Text != "Society & Culture" {
		t.Errorf("Expected category 'Society & Culture', got '%s'", cat.Text)
	}

	if len(cat.ItunesCategories) != 1 {
		t.Errorf("Expected 1 subcategory, got %d", len(cat.ItunesCategories))
	}

	if cat.ItunesCategories[0].Text != "Documentary" {
		t.Errorf("Expected subcategory 'Documentary', got '%s'", cat.ItunesCategories[0].Text)
	}

	// Test multiple categories
	channelDataMultiple := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories: []Category{
			NewCategory("Technology"),
			NewCategory("Business", "Entrepreneurship"),
		},
	}

	feedMultiple := NewFeed(channelDataMultiple)
	if len(feedMultiple.xmlDoc.Channel.ItunesCategory) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(feedMultiple.xmlDoc.Channel.ItunesCategory))
	}
}

func TestFeedAddItem(t *testing.T) {
	channelData := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories:  []Category{NewCategory("Technology")},
	}

	feed := NewFeed(channelData)

	// Initially no items
	if len(feed.xmlDoc.Channel.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(feed.xmlDoc.Channel.Items))
	}

	// Add an item
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024,
			Type:   Mp3,
		},
	}

	item := NewItem(itemData)
	feed.AddItem(item)

	if len(feed.xmlDoc.Channel.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(feed.xmlDoc.Channel.Items))
	}

	addedItem := feed.xmlDoc.Channel.Items[0]
	if addedItem.Title != itemData.Title {
		t.Errorf("Expected item title %s, got %s", itemData.Title, addedItem.Title)
	}

	if addedItem.Guid != itemData.Guid {
		t.Errorf("Expected item Guid %s, got %s", itemData.Guid, addedItem.Guid)
	}
}

func TestFeedXMLGeneration(t *testing.T) {
	channelData := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description with <special> characters & ampersands",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories:  []Category{NewCategory("Technology")},
	}

	feed := NewFeed(channelData)

	// Add a valid item to pass validation
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024,
			Type:   Mp3,
		},
	}
	item := NewItem(itemData)
	feed.AddItem(item)

	var buf bytes.Buffer
	err := feed.Encode(&buf)
	if err != nil {
		t.Fatalf("Failed to encode feed: %v", err)
	}

	xmlContent := buf.String()

	// Check XML header
	if !strings.HasPrefix(xmlContent, xml.Header) {
		t.Error("XML should start with XML header")
	}

	// Check RSS version
	if !strings.Contains(xmlContent, `version="2.0"`) {
		t.Error("RSS version should be 2.0")
	}

	// Check namespaces
	if !strings.Contains(xmlContent, `xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"`) {
		t.Error("iTunes namespace should be present")
	}

	if !strings.Contains(xmlContent, `xmlns:content="http://purl.org/rss/1.0/modules/content/"`) {
		t.Error("Content namespace should be present")
	}

	if !strings.Contains(xmlContent, `xmlns:podcast="https://podcastindex.org/namespace/1.0"`) {
		t.Error("Podcast namespace should be present")
	}

	// Check CDATA handling
	if !strings.Contains(xmlContent, "<![CDATA[A test podcast description with <special> characters & ampersands]]>") {
		t.Error("Description should be wrapped in CDATA")
	}

	// Parse the generated XML to ensure it's valid
	var doc xmlDoc
	err = xml.Unmarshal(buf.Bytes(), &doc)
	if err != nil {
		t.Fatalf("Generated XML is not valid: %v", err)
	}

	// Verify parsed content matches original
	if doc.Channel.Title != channelData.Title {
		t.Errorf("Parsed title doesn't match: expected %s, got %s", channelData.Title, doc.Channel.Title)
	}
}

func TestFeedEncodeValidation(t *testing.T) {
	// Test that Encode fails for invalid feeds
	channelData := FeedData{
		Title:       "Test Podcast",
		Description: "A test podcast description",
		Image:       "https://example.com/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories:  []Category{NewCategory("Technology")},
	}

	feed := NewFeed(channelData)
	// Don't add any items - this should fail validation

	var buf bytes.Buffer
	err := feed.Encode(&buf)
	if err == nil {
		t.Error("Expected encode to fail for feed without items")
	} else {
		if !strings.Contains(err.Error(), "feed validation failed") {
			t.Errorf("Expected validation error, got: %v", err)
		}
	}
}
