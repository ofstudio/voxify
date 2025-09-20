package feedcast

import (
	"strings"
	"testing"
	"time"
)

func TestNewItem(t *testing.T) {
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024000,
			Type:   Mp3,
		},
	}

	item := NewItem(itemData)

	if item == nil {
		t.Fatal("NewItem returned nil")
	}

	// Test basic structure
	if item.xmlItem.Title != itemData.Title {
		t.Errorf("Expected title %s, got %s", itemData.Title, item.xmlItem.Title)
	}

	if item.xmlItem.Guid != itemData.Guid {
		t.Errorf("Expected Guid %s, got %s", itemData.Guid, item.xmlItem.Guid)
	}

	if item.xmlItem.Enclosure.URL != itemData.Enclosure.URL {
		t.Errorf("Expected enclosure URL %s, got %s", itemData.Enclosure.URL, item.xmlItem.Enclosure.URL)
	}

	if item.xmlItem.Enclosure.Length != itemData.Enclosure.Length {
		t.Errorf("Expected enclosure length %d, got %d", itemData.Enclosure.Length, item.xmlItem.Enclosure.Length)
	}

	if item.xmlItem.Enclosure.Type != itemData.Enclosure.Type {
		t.Errorf("Expected enclosure type %s, got %s", itemData.Enclosure.Type, item.xmlItem.Enclosure.Type)
	}
}

func TestItemValidation_RequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		itemData    ItemData
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid item",
			itemData: ItemData{
				Title: "Valid Episode",
				Guid:  "valid-episode-1",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: 1024000,
					Type:   Mp3,
				},
			},
			expectError: false,
		},
		{
			name: "missing title",
			itemData: ItemData{
				Title: "",
				Guid:  "valid-episode-1",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: 1024000,
					Type:   Mp3,
				},
			},
			expectError: true,
			errorMsg:    "item title is required",
		},
		{
			name: "missing Guid",
			itemData: ItemData{
				Title: "Valid Episode",
				Guid:  "",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: 1024000,
					Type:   Mp3,
				},
			},
			expectError: true,
			errorMsg:    "item guid is required",
		},
		{
			name: "missing enclosure URL",
			itemData: ItemData{
				Title: "Valid Episode",
				Guid:  "valid-episode-1",
				Enclosure: Enclosure{
					URL:    "",
					Length: 1024000,
					Type:   Mp3,
				},
			},
			expectError: true,
			errorMsg:    "item enclosure url is required",
		},
		{
			name: "zero enclosure length",
			itemData: ItemData{
				Title: "Valid Episode",
				Guid:  "valid-episode-1",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: 0,
					Type:   Mp3,
				},
			},
			expectError: true,
			errorMsg:    "item enclosure length must be greater than zero",
		},
		{
			name: "negative enclosure length",
			itemData: ItemData{
				Title: "Valid Episode",
				Guid:  "valid-episode-1",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: -1,
					Type:   Mp3,
				},
			},
			expectError: true,
			errorMsg:    "item enclosure length must be greater than zero",
		},
		{
			name: "missing enclosure type",
			itemData: ItemData{
				Title: "Valid Episode",
				Guid:  "valid-episode-1",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: 1024000,
					Type:   "",
				},
			},
			expectError: true,
			errorMsg:    "item enclosure type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.itemData)
			err := item.Validate()

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

func TestItemWithOptionalFields(t *testing.T) {
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024000,
			Type:   Mp3,
		},
	}

	item := NewItem(itemData)

	// Test WithPubDate
	pubDate := time.Date(2023, 4, 1, 19, 0, 0, 0, time.UTC)
	item.WithPubDate(pubDate)
	expectedPubDate := pubDate.Format(time.RFC1123Z)
	if item.xmlItem.PubDate != expectedPubDate {
		t.Errorf("Expected pubDate %s, got %s", expectedPubDate, item.xmlItem.PubDate)
	}

	// Test WithDescription
	description := "This is a test episode description with <html> tags & special characters"
	item.WithDescription(description)
	if item.xmlItem.Description == nil || item.xmlItem.Description.Data != description {
		t.Errorf("Expected description %s, got %v", description, item.xmlItem.Description)
	}

	// Test WithItunesDuration
	duration := int64(3600) // 1 hour in seconds
	item.WithItunesDuration(duration)
	if item.xmlItem.ItunesDuration != "3600" {
		t.Errorf("Expected duration '3600', got '%s'", item.xmlItem.ItunesDuration)
	}

	// Test WithLink
	link := "https://example.com/episodes/episode1"
	item.WithLink(link)
	if item.xmlItem.Link != link {
		t.Errorf("Expected link %s, got %s", link, item.xmlItem.Link)
	}

	// Test WithItunesImage
	imageURL := "https://example.com/episode1-artwork.jpg"
	item.WithItunesImage(imageURL)
	if item.xmlItem.ItunesImage == nil || item.xmlItem.ItunesImage.Href != imageURL {
		t.Errorf("Expected iTunes image %s, got %v", imageURL, item.xmlItem.ItunesImage)
	}

	// Test WithItunesExplicit
	item.WithItunesExplicit(ExplicitTrue)
	if item.xmlItem.ItunesExplicit != ExplicitTrue {
		t.Errorf("Expected iTunes explicit %s, got %s", ExplicitTrue, item.xmlItem.ItunesExplicit)
	}

	// Test WithItunesTitle
	itunesTitle := "Episode 1: Introduction"
	item.WithItunesTitle(itunesTitle)
	if item.xmlItem.ItunesTitle != itunesTitle {
		t.Errorf("Expected iTunes title %s, got %s", itunesTitle, item.xmlItem.ItunesTitle)
	}

	// Test WithItunesEpisode
	episodeNumber := 1
	item.WithItunesEpisode(episodeNumber)
	if item.xmlItem.ItunesEpisode != "1" {
		t.Errorf("Expected iTunes episode '1', got '%s'", item.xmlItem.ItunesEpisode)
	}

	// Test WithItunesEpisode with zero (should not set)
	item.WithItunesEpisode(0)
	if item.xmlItem.ItunesEpisode != "1" { // Should remain unchanged
		t.Errorf("Expected iTunes episode to remain '1', got '%s'", item.xmlItem.ItunesEpisode)
	}

	// Test WithItunesSeason
	seasonNumber := 1
	item.WithItunesSeason(seasonNumber)
	if item.xmlItem.ItunesSeason != "1" {
		t.Errorf("Expected iTunes season '1', got '%s'", item.xmlItem.ItunesSeason)
	}

	// Test WithItunesSeason with zero (should not set)
	item.WithItunesSeason(0)
	if item.xmlItem.ItunesSeason != "1" { // Should remain unchanged
		t.Errorf("Expected iTunes season to remain '1', got '%s'", item.xmlItem.ItunesSeason)
	}

	// Test WithItunesEpisodeType
	item.WithItunesEpisodeType(EpisodeFull)
	if item.xmlItem.ItunesEpisodeType != EpisodeFull {
		t.Errorf("Expected iTunes episode type %s, got %s", EpisodeFull, item.xmlItem.ItunesEpisodeType)
	}

	// Test WithPodcastTranscript
	transcriptURL := "https://example.com/transcripts/episode1.vtt"
	item.WithPodcastTranscript(transcriptURL, TranscriptVtt)
	if len(item.xmlItem.PodcastTranscripts) != 1 {
		t.Errorf("Expected 1 transcript, got %d", len(item.xmlItem.PodcastTranscripts))
	} else {
		transcript := item.xmlItem.PodcastTranscripts[0]
		if transcript.Url != transcriptURL {
			t.Errorf("Expected transcript URL %s, got %s", transcriptURL, transcript.Url)
		}
		if transcript.Type != string(TranscriptVtt) {
			t.Errorf("Expected transcript type %s, got %s", TranscriptVtt, transcript.Type)
		}
	}

	// Test multiple transcripts
	srtURL := "https://example.com/transcripts/episode1.srt"
	item.WithPodcastTranscript(srtURL, TranscriptSrt)
	if len(item.xmlItem.PodcastTranscripts) != 2 {
		t.Errorf("Expected 2 transcripts, got %d", len(item.xmlItem.PodcastTranscripts))
	}

	// Test WithItunesBlock
	item.WithItunesBlock(BlockYes)
	if item.xmlItem.ItunesBlock != BlockYes {
		t.Errorf("Expected iTunes block %s, got %s", BlockYes, item.xmlItem.ItunesBlock)
	}

	// Test WithItunesAuthor
	author := "Episode Author"
	item.WithItunesAuthor(author)
	if item.xmlItem.ItunesAuthor != author {
		t.Errorf("Expected iTunes author %s, got %s", author, item.xmlItem.ItunesAuthor)
	}

	// Test WithItunesSummary
	summary := "This is an episode summary with <formatting> & special characters"
	item.WithItunesSummary(summary)
	if item.xmlItem.ItunesSummary == nil || item.xmlItem.ItunesSummary.Data != summary {
		t.Errorf("Expected iTunes summary %s, got %v", summary, item.xmlItem.ItunesSummary)
	}
}

func TestItemExplicitValidation(t *testing.T) {
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024000,
			Type:   Mp3,
		},
	}

	tests := []struct {
		name        string
		explicit    Explicit
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid explicit true",
			explicit:    ExplicitTrue,
			expectError: false,
		},
		{
			name:        "valid explicit false",
			explicit:    ExplicitFalse,
			expectError: false,
		},
		{
			name:        "empty explicit (valid)",
			explicit:    ExplicitNotSet,
			expectError: false,
		},
		{
			name:        "invalid explicit value",
			explicit:    "invalid",
			expectError: true,
			errorMsg:    "item itunes:explicit must be either 'true' or 'false'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(itemData)
			item.WithItunesExplicit(tt.explicit)
			err := item.Validate()

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

func TestItemEnclosureTypes(t *testing.T) {
	validEnclosureTypes := []EnclosureType{
		M4a,
		Mp3,
		Mov,
		Mp4,
		M4v,
		Pdf,
	}

	for _, encType := range validEnclosureTypes {
		t.Run(string(encType), func(t *testing.T) {
			itemData := ItemData{
				Title: "Test Episode",
				Guid:  "test-episode-1",
				Enclosure: Enclosure{
					URL:    "https://example.com/episode1.mp3",
					Length: 1024000,
					Type:   encType,
				},
			}

			item := NewItem(itemData)
			err := item.Validate()
			if err != nil {
				t.Errorf("Valid enclosure type %s should not cause validation error: %v", encType, err)
			}

			if item.xmlItem.Enclosure.Type != encType {
				t.Errorf("Expected enclosure type %s, got %s", encType, item.xmlItem.Enclosure.Type)
			}
		})
	}
}

func TestItemEpisodeTypes(t *testing.T) {
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024000,
			Type:   Mp3,
		},
	}

	episodeTypes := []ItunesEpisodeType{
		EpisodeFull,
		EpisodeTrailer,
		EpisodeBonus,
	}

	for _, epType := range episodeTypes {
		t.Run(string(epType), func(t *testing.T) {
			item := NewItem(itemData)
			item.WithItunesEpisodeType(epType)

			if item.xmlItem.ItunesEpisodeType != epType {
				t.Errorf("Expected episode type %s, got %s", epType, item.xmlItem.ItunesEpisodeType)
			}

			err := item.Validate()
			if err != nil {
				t.Errorf("Valid episode type %s should not cause validation error: %v", epType, err)
			}
		})
	}
}

func TestItemTranscriptTypes(t *testing.T) {
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024000,
			Type:   Mp3,
		},
	}

	transcriptTypes := []PodcastTranscriptType{
		TranscriptVtt,
		TranscriptSrt,
		TranscriptSubrip,
	}

	for _, transType := range transcriptTypes {
		t.Run(string(transType), func(t *testing.T) {
			item := NewItem(itemData)
			transcriptURL := "https://example.com/transcript.vtt"
			item.WithPodcastTranscript(transcriptURL, transType)

			if len(item.xmlItem.PodcastTranscripts) != 1 {
				t.Errorf("Expected 1 transcript, got %d", len(item.xmlItem.PodcastTranscripts))
			} else {
				transcript := item.xmlItem.PodcastTranscripts[0]
				if transcript.Type != string(transType) {
					t.Errorf("Expected transcript type %s, got %s", transType, transcript.Type)
				}
			}

			err := item.Validate()
			if err != nil {
				t.Errorf("Valid transcript type %s should not cause validation error: %v", transType, err)
			}
		})
	}
}

func TestItemFluentAPI(t *testing.T) {
	itemData := ItemData{
		Title: "Test Episode",
		Guid:  "test-episode-1",
		Enclosure: Enclosure{
			URL:    "https://example.com/episode1.mp3",
			Length: 1024000,
			Type:   Mp3,
		},
	}

	// Test that all With* methods return the item for chaining
	pubDate := time.Date(2023, 4, 1, 19, 0, 0, 0, time.UTC)

	item := NewItem(itemData).
		WithPubDate(pubDate).
		WithDescription("Test description").
		WithItunesDuration(3600).
		WithLink("https://example.com/episode1").
		WithItunesImage("https://example.com/image.jpg").
		WithItunesExplicit(ExplicitFalse).
		WithItunesTitle("Episode Title").
		WithItunesEpisode(1).
		WithItunesSeason(1).
		WithItunesEpisodeType(EpisodeFull).
		WithPodcastTranscript("https://example.com/transcript.vtt", TranscriptVtt).
		WithItunesBlock(BlockNotSet).
		WithItunesAuthor("Author Name").
		WithItunesSummary("Episode summary")

	if item == nil {
		t.Fatal("Fluent API chain returned nil")
	}

	// Verify all fields were set
	if item.xmlItem.PubDate != pubDate.Format(time.RFC1123Z) {
		t.Error("PubDate not set correctly in fluent chain")
	}

	if item.xmlItem.Description == nil || item.xmlItem.Description.Data != "Test description" {
		t.Error("Description not set correctly in fluent chain")
	}

	if item.xmlItem.ItunesDuration != "3600" {
		t.Error("Duration not set correctly in fluent chain")
	}

	if item.xmlItem.ItunesEpisode != "1" {
		t.Error("Episode number not set correctly in fluent chain")
	}

	if len(item.xmlItem.PodcastTranscripts) != 1 {
		t.Error("Transcript not added correctly in fluent chain")
	}

	// Verify the item validates
	err := item.Validate()
	if err != nil {
		t.Errorf("Fluent API item should validate successfully: %v", err)
	}
}
