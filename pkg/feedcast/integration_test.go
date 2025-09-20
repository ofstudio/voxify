package feedcast

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

// TestApplePodcastCompliance_FullFeed tests a complete RSS feed
// against Apple Podcast requirements with real-world data
func TestApplePodcastCompliance_FullFeed(t *testing.T) {
	// Create a comprehensive feed that meets all Apple requirements
	channelData := FeedData{
		Title:       "Tech Talk Weekly",
		Description: "The latest in technology news, reviews, and interviews with industry leaders. Join us every week for deep dives into emerging technologies, startup stories, and expert analysis.",
		Image:       "https://example.com/podcasts/tech-talk-weekly/artwork.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories: []Category{
			NewCategory("Technology"),
			NewCategory("Business", "Entrepreneurship"),
		},
	}

	feed := NewFeed(channelData)

	// Add optional channel fields
	feed.WithAuthor("TechMedia Network").
		WithLink("https://techtalkweekly.com").
		WithPubDate(time.Date(2023, 4, 15, 12, 0, 0, 0, time.UTC)).
		WithLastBuildDate(time.Date(2023, 4, 15, 12, 30, 0, 0, time.UTC)).
		WithItunesTitle("Tech Talk Weekly").
		WithItunesType(TypeEpisodic).
		WithCopyright("© 2023 TechMedia Network").
		WithGenerator("Voxify RSS Generator v1.0").
		WithItunesSummary("The premier technology podcast for professionals and enthusiasts. We cover everything from AI and machine learning to startup funding and product launches.").
		WithItunesKeywords("technology,business,startups,AI,software,hardware").
		WithItunesOwner("John Smith", "john@techtalkweekly.com")

	// Add multiple episodes with different configurations
	episodes := []struct {
		title         string
		guid          string
		description   string
		duration      int64
		episodeNum    int
		season        int
		explicit      Explicit
		episodeType   ItunesEpisodeType
		hasImage      bool
		hasTranscript bool
	}{
		{
			title:         "The Future of Artificial Intelligence",
			guid:          "tti-2023-001",
			description:   "We dive deep into the latest developments in AI, including <strong>GPT-4</strong>, autonomous vehicles, and the ethical implications of artificial intelligence in society.",
			duration:      3600, // 1 hour
			episodeNum:    1,
			season:        1,
			explicit:      ExplicitFalse,
			episodeType:   EpisodeFull,
			hasImage:      true,
			hasTranscript: true,
		},
		{
			title:         "Season 1 Trailer",
			guid:          "tti-2023-trailer",
			description:   "Get ready for Season 1 of Tech Talk Weekly! Here's what to expect in our upcoming episodes.",
			duration:      120, // 2 minutes
			episodeNum:    0,   // No episode number for trailer
			season:        1,
			explicit:      ExplicitFalse,
			episodeType:   EpisodeTrailer,
			hasImage:      false,
			hasTranscript: false,
		},
		{
			title:         "Startup Funding in 2023: What Investors Want",
			guid:          "tti-2023-002",
			description:   "Join us as we interview three successful venture capitalists about the current funding landscape. <em>Warning: Contains frank discussions about business failures.</em>",
			duration:      2700, // 45 minutes
			episodeNum:    2,
			season:        1,
			explicit:      ExplicitTrue, // Due to frank business discussions
			episodeType:   EpisodeFull,
			hasImage:      true,
			hasTranscript: true,
		},
		{
			title:         "Behind the Scenes: How We Make the Show",
			guid:          "tti-2023-bonus-001",
			description:   "A special bonus episode where we take you behind the scenes of our production process.",
			duration:      1800, // 30 minutes
			episodeNum:    0,    // No episode number for bonus
			season:        1,
			explicit:      ExplicitFalse,
			episodeType:   EpisodeBonus,
			hasImage:      false,
			hasTranscript: false,
		},
	}

	for i, ep := range episodes {
		itemData := ItemData{
			Title: ep.title,
			Guid:  ep.guid,
			Enclosure: Enclosure{
				URL:    "https://cdn.techtalkweekly.com/episodes/" + ep.guid + ".mp3",
				Length: 25000000 + int64(i*1000000), // Varying file sizes
				Type:   Mp3,
			},
		}

		item := NewItem(itemData).
			WithPubDate(time.Date(2023, 4, 1+i*7, 12, 0, 0, 0, time.UTC)).
			WithDescription(ep.description).
			WithItunesDuration(ep.duration).
			WithLink("https://techtalkweekly.com/episodes/" + ep.guid).
			WithItunesExplicit(ep.explicit).
			WithItunesTitle(ep.title).
			WithItunesEpisodeType(ep.episodeType).
			WithItunesAuthor("TechMedia Network")

		if ep.episodeNum > 0 {
			item.WithItunesEpisode(ep.episodeNum)
		}

		if ep.season > 0 {
			item.WithItunesSeason(ep.season)
		}

		if ep.hasImage {
			item.WithItunesImage("https://cdn.techtalkweekly.com/episodes/" + ep.guid + "-artwork.jpg")
		}

		if ep.hasTranscript {
			item.WithPodcastTranscript("https://cdn.techtalkweekly.com/transcripts/"+ep.guid+".vtt", TranscriptVtt).
				WithPodcastTranscript("https://cdn.techtalkweekly.com/transcripts/"+ep.guid+".srt", TranscriptSrt)
		}

		feed.AddItem(item)
	}

	// Test validation
	err := feed.Validate()
	if err != nil {
		t.Fatalf("Feed validation failed: %v", err)
	}

	// Test XML generation
	var buf bytes.Buffer
	err = feed.Encode(&buf)
	if err != nil {
		t.Fatalf("Feed encoding failed: %v", err)
	}

	xmlContent := buf.String()

	// Verify XML structure and Apple Podcast compliance
	testApplePodcastXMLCompliance(t, xmlContent)

	// Parse and verify the generated XML
	var parsedDoc xmlDoc
	err = xml.Unmarshal(buf.Bytes(), &parsedDoc)
	if err != nil {
		t.Fatalf("Generated XML is not valid: %v", err)
	}

	// Verify parsed content matches expectations
	testParsedFeedCompliance(t, parsedDoc, channelData)
}

func testApplePodcastXMLCompliance(t *testing.T, xmlContent string) {
	t.Helper()

	// Test XML structure requirements
	requiredElements := []string{
		xml.Header,
		`<rss version="2.0"`,
		`xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"`,
		`xmlns:content="http://purl.org/rss/1.0/modules/content/"`,
		`xmlns:podcast="https://podcastindex.org/namespace/1.0"`,
		"<channel>",
		"<title>",
		"<description>",
		"<itunes:image",
		"<language>",
		"<itunes:explicit>",
		"<itunes:category",
		"<item>",
		"<guid>",
		"<enclosure",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(xmlContent, elem) {
			t.Errorf("Required element missing from XML: %s", elem)
		}
	}

	// Test CDATA usage for descriptions
	if !strings.Contains(xmlContent, "<![CDATA[") {
		t.Error("CDATA sections should be used for descriptions")
	}

	// Test proper namespace declarations
	if !strings.Contains(xmlContent, `xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"`) {
		t.Error("iTunes namespace declaration missing or incorrect")
	}

	if !strings.Contains(xmlContent, `xmlns:podcast="https://podcastindex.org/namespace/1.0"`) {
		t.Error("Podcast namespace declaration missing or incorrect")
	}

	// Test RSS version
	if !strings.Contains(xmlContent, `version="2.0"`) {
		t.Error("RSS version should be 2.0")
	}

	// Test episode-specific elements
	episodeElements := []string{
		"<itunes:duration>",
		"<itunes:episode>",
		"<itunes:season>",
		"<itunes:episodeType>",
		"<podcast:transcript",
	}

	for _, elem := range episodeElements {
		if !strings.Contains(xmlContent, elem) {
			t.Errorf("Episode element missing from XML: %s", elem)
		}
	}

	// Test proper attribute formatting
	attributeTests := []string{
		`href="https://`,    // iTunes image
		`url="https://`,     // Enclosure and transcript URLs
		`length="`,          // Enclosure length
		`type="audio/mpeg"`, // Enclosure type
		`text="Technology"`, // Category text
	}

	for _, attr := range attributeTests {
		if !strings.Contains(xmlContent, attr) {
			t.Errorf("Required attribute format missing from XML: %s", attr)
		}
	}
}

func testParsedFeedCompliance(t *testing.T, doc xmlDoc, channelData FeedData) {
	t.Helper()

	// Test channel compliance
	channel := doc.Channel

	if channel.Title != channelData.Title {
		t.Errorf("Channel title mismatch: expected %s, got %s", channelData.Title, channel.Title)
	}

	if channel.Description.Data != channelData.Description {
		t.Errorf("Channel description mismatch: expected %s, got %s", channelData.Description, channel.Description.Data)
	}

	// Note: XML namespace parsing can be tricky, so we'll validate the XML string content instead
	// of relying on the parsed struct for namespace-prefixed fields

	if channel.Language != channelData.Language {
		t.Errorf("Channel language mismatch: expected %s, got %s", channelData.Language, channel.Language)
	}

	// Test items compliance - basic structure
	if len(channel.Items) != 4 { // We added 4 episodes
		t.Errorf("Item count mismatch: expected 4, got %d", len(channel.Items))
	}

	// Test specific episode requirements
	for i, item := range channel.Items {
		// Every item must have required fields
		if item.Title == "" {
			t.Errorf("Item %d missing title", i)
		}
		if item.Guid == "" {
			t.Errorf("Item %d missing Guid", i)
		}
		if item.Enclosure.URL == "" {
			t.Errorf("Item %d missing enclosure URL", i)
		}
		if item.Enclosure.Length <= 0 {
			t.Errorf("Item %d has invalid enclosure length: %d", i, item.Enclosure.Length)
		}
		if item.Enclosure.Type == "" {
			t.Errorf("Item %d missing enclosure type", i)
		}
	}

	// For namespace-dependent fields, we'll validate them in the XML compliance test
	// rather than the parsed struct test, since XML namespace handling is complex
}

// TestApplePodcastCompliance_MinimalFeed tests the absolute minimum
// required for Apple Podcast compliance
func TestApplePodcastCompliance_MinimalFeed(t *testing.T) {
	channelData := FeedData{
		Title:       "Minimal Podcast",
		Description: "Minimal description",
		Image:       "https://example.com/minimal.jpg",
		Language:    "en",
		Explicit:    ExplicitFalse,
		Categories:  []Category{NewCategory("Education")},
	}

	feed := NewFeed(channelData)

	// Add one minimal episode
	itemData := ItemData{
		Title: "Episode 1",
		Guid:  "ep001",
		Enclosure: Enclosure{
			URL:    "https://example.com/ep001.mp3",
			Length: 1000000, // 1MB
			Type:   Mp3,
		},
	}

	item := NewItem(itemData)
	feed.AddItem(item)

	// This should validate successfully
	err := feed.Validate()
	if err != nil {
		t.Fatalf("Minimal feed should validate: %v", err)
	}

	// Should encode without errors
	var buf bytes.Buffer
	err = feed.Encode(&buf)
	if err != nil {
		t.Fatalf("Minimal feed should encode: %v", err)
	}

	xmlContent := buf.String()

	// Should contain all required elements
	requiredForMinimal := []string{
		"<title>Minimal Podcast</title>",
		"<itunes:explicit>false</itunes:explicit>",
		"<language>en</language>",
		`<itunes:category text="Education"`,
		"<item>",
		"<title>Episode 1</title>",
		"<guid>ep001</guid>",
		`type="audio/mpeg"`,
	}

	for _, req := range requiredForMinimal {
		if !strings.Contains(xmlContent, req) {
			t.Errorf("Minimal feed missing required element: %s", req)
		}
	}
}

// TestApplePodcastCompliance_SerialPodcast tests compliance for serial podcasts
func TestApplePodcastCompliance_SerialPodcast(t *testing.T) {
	channelData := FeedData{
		Title:       "Mystery Serial",
		Description: "A thrilling mystery told across multiple episodes",
		Image:       "https://example.com/mystery.jpg",
		Language:    "en",
		Explicit:    ExplicitTrue, // Contains mature themes
		Categories:  []Category{NewCategory("Fiction", "Drama")},
	}

	feed := NewFeed(channelData)
	feed.WithItunesType(TypeSerial) // Serial shows require episode numbers

	// Add episodes in order (serial shows should have episode numbers)
	for i := 1; i <= 3; i++ {
		itemData := ItemData{
			Title: "Episode " + string(rune('0'+i)),
			Guid:  "mystery-ep-" + string(rune('0'+i)),
			Enclosure: Enclosure{
				URL:    "https://example.com/mystery/ep" + string(rune('0'+i)) + ".mp3",
				Length: int64(20000000 + i*1000000),
				Type:   Mp3,
			},
		}

		item := NewItem(itemData).
			WithItunesEpisode(i). // Required for serial shows
			WithItunesSeason(1).  // Optional but recommended
			WithItunesExplicit(ExplicitTrue)

		feed.AddItem(item)
	}

	err := feed.Validate()
	if err != nil {
		t.Fatalf("Serial feed should validate: %v", err)
	}

	var buf bytes.Buffer
	err = feed.Encode(&buf)
	if err != nil {
		t.Fatalf("Serial feed should encode: %v", err)
	}

	xmlContent := buf.String()

	// Verify serial-specific requirements
	if !strings.Contains(xmlContent, "<itunes:type>serial</itunes:type>") {
		t.Error("Serial feed should contain itunes:type")
	}

	// Should have episode numbers
	for i := 1; i <= 3; i++ {
		episodeTag := "<itunes:episode>" + string(rune('0'+i)) + "</itunes:episode>"
		if !strings.Contains(xmlContent, episodeTag) {
			t.Errorf("Serial feed should contain episode number: %s", episodeTag)
		}
	}
}

// TestApplePodcastCompliance_ErrorCases tests various error conditions
func TestApplePodcastCompliance_ErrorCases(t *testing.T) {
	errorTests := []struct {
		name        string
		setupFeed   func() *Feed
		expectedErr string
	}{
		{
			name: "missing title",
			setupFeed: func() *Feed {
				channelData := FeedData{
					Title:       "", // Missing title
					Description: "Valid description",
					Image:       "https://example.com/image.jpg",
					Language:    "en",
					Explicit:    ExplicitFalse,
					Categories:  []Category{NewCategory("Technology")},
				}
				return NewFeed(channelData)
			},
			expectedErr: "channel title is required",
		},
		{
			name: "invalid explicit value",
			setupFeed: func() *Feed {
				channelData := FeedData{
					Title:       "Test Podcast",
					Description: "Valid description",
					Image:       "https://example.com/image.jpg",
					Language:    "en",
					Explicit:    "maybe", // Invalid explicit
					Categories:  []Category{NewCategory("Technology")},
				}
				return NewFeed(channelData)
			},
			expectedErr: "channel itunes:explicit must be either 'true' or 'false'",
		},
		{
			name: "no categories",
			setupFeed: func() *Feed {
				channelData := FeedData{
					Title:       "Test Podcast",
					Description: "Valid description",
					Image:       "https://example.com/image.jpg",
					Language:    "en",
					Explicit:    ExplicitFalse,
					Categories:  []Category{}, // No categories
				}
				return NewFeed(channelData)
			},
			expectedErr: "at least one itunes:category is required",
		},
		{
			name: "no episodes",
			setupFeed: func() *Feed {
				channelData := FeedData{
					Title:       "Test Podcast",
					Description: "Valid description",
					Image:       "https://example.com/image.jpg",
					Language:    "en",
					Explicit:    ExplicitFalse,
					Categories:  []Category{NewCategory("Technology")},
				}
				return NewFeed(channelData)
				// Don't add any episodes
			},
			expectedErr: "at least one channel item is required",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			feed := tt.setupFeed()
			err := feed.Validate()

			if err == nil {
				t.Error("Expected validation to fail but it passed")
			} else if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}
