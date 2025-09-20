package feedcast

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestXmlDocValidation(t *testing.T) {
	tests := []struct {
		name        string
		doc         xmlDoc
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid xml doc",
			doc: xmlDoc{
				Version:   xmlRssVersion,
				ItunesNS:  xmlItunesNS,
				ContentNS: xmlContentNS,
				PodcastNS: xmlPodcastNS,
				Channel: xmlChannel{
					Title:          "Valid Podcast",
					Description:    xmlCDATA{Data: "Valid description"},
					ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
					Language:       "en",
					ItunesExplicit: ExplicitFalse,
					ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
					Items: []xmlItem{{
						Title:     "Test Episode",
						Guid:      "test-guid",
						Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
					}},
				},
			},
			expectError: false,
		},
		{
			name: "invalid RSS version",
			doc: xmlDoc{
				Version:   "1.0",
				ItunesNS:  xmlItunesNS,
				ContentNS: xmlContentNS,
				PodcastNS: xmlPodcastNS,
				Channel: xmlChannel{
					Title:          "Valid Podcast",
					Description:    xmlCDATA{Data: "Valid description"},
					ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
					Language:       "en",
					ItunesExplicit: ExplicitFalse,
					ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
					Items: []xmlItem{{
						Title:     "Test Episode",
						Guid:      "test-guid",
						Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
					}},
				},
			},
			expectError: true,
			errorMsg:    "feed version mismatch",
		},
		{
			name: "invalid iTunes namespace",
			doc: xmlDoc{
				Version:   xmlRssVersion,
				ItunesNS:  "invalid-namespace",
				ContentNS: xmlContentNS,
				PodcastNS: xmlPodcastNS,
				Channel: xmlChannel{
					Title:          "Valid Podcast",
					Description:    xmlCDATA{Data: "Valid description"},
					ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
					Language:       "en",
					ItunesExplicit: ExplicitFalse,
					ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
					Items: []xmlItem{{
						Title:     "Test Episode",
						Guid:      "test-guid",
						Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
					}},
				},
			},
			expectError: true,
			errorMsg:    "itunes namespace mismatch",
		},
		{
			name: "invalid content namespace",
			doc: xmlDoc{
				Version:   xmlRssVersion,
				ItunesNS:  xmlItunesNS,
				ContentNS: "invalid-namespace",
				PodcastNS: xmlPodcastNS,
				Channel: xmlChannel{
					Title:          "Valid Podcast",
					Description:    xmlCDATA{Data: "Valid description"},
					ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
					Language:       "en",
					ItunesExplicit: ExplicitFalse,
					ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
					Items: []xmlItem{{
						Title:     "Test Episode",
						Guid:      "test-guid",
						Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
					}},
				},
			},
			expectError: true,
			errorMsg:    "content namespace mismatch",
		},
		{
			name: "invalid podcast namespace",
			doc: xmlDoc{
				Version:   xmlRssVersion,
				ItunesNS:  xmlItunesNS,
				ContentNS: xmlContentNS,
				PodcastNS: "invalid-namespace",
				Channel: xmlChannel{
					Title:          "Valid Podcast",
					Description:    xmlCDATA{Data: "Valid description"},
					ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
					Language:       "en",
					ItunesExplicit: ExplicitFalse,
					ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
					Items: []xmlItem{{
						Title:     "Test Episode",
						Guid:      "test-guid",
						Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
					}},
				},
			},
			expectError: true,
			errorMsg:    "podcast namespace mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.doc.validate()

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

func TestXmlChannelValidation(t *testing.T) {
	validItem := xmlItem{
		Title:     "Test Episode",
		Guid:      "test-guid",
		Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
	}

	tests := []struct {
		name        string
		channel     xmlChannel
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid channel",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{validItem},
			},
			expectError: false,
		},
		{
			name: "empty title",
			channel: xmlChannel{
				Title:          "",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{validItem},
			},
			expectError: true,
			errorMsg:    "channel title is required",
		},
		{
			name: "empty description",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: ""},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{validItem},
			},
			expectError: true,
			errorMsg:    "channel description is required",
		},
		{
			name: "empty iTunes image",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: ""},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{validItem},
			},
			expectError: true,
			errorMsg:    "channel itunes:image is required",
		},
		{
			name: "empty language",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{validItem},
			},
			expectError: true,
			errorMsg:    "channel language is required",
		},
		{
			name: "invalid explicit value",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: "invalid",
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{validItem},
			},
			expectError: true,
			errorMsg:    "channel itunes:explicit must be either 'true' or 'false'",
		},
		{
			name: "no categories",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{},
				Items:          []xmlItem{validItem},
			},
			expectError: true,
			errorMsg:    "at least one itunes:category is required",
		},
		{
			name: "no items",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items:          []xmlItem{},
			},
			expectError: true,
			errorMsg:    "at least one channel item is required",
		},
		{
			name: "invalid item",
			channel: xmlChannel{
				Title:          "Valid Podcast",
				Description:    xmlCDATA{Data: "Valid description"},
				ItunesImage:    xmlItunesImage{Href: "https://example.com/image.jpg"},
				Language:       "en",
				ItunesExplicit: ExplicitFalse,
				ItunesCategory: []xmlItunesCategory{{Text: "Technology"}},
				Items: []xmlItem{{
					Title:     "", // Invalid: empty title
					Guid:      "test-guid",
					Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
				}},
			},
			expectError: true,
			errorMsg:    "invalid item 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.channel.validate()

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

func TestXmlItemValidation(t *testing.T) {
	tests := []struct {
		name        string
		item        xmlItem
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid item",
			item: xmlItem{
				Title:     "Valid Episode",
				Guid:      "valid-guid",
				Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
			},
			expectError: false,
		},
		{
			name: "empty title",
			item: xmlItem{
				Title:     "",
				Guid:      "valid-guid",
				Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
			},
			expectError: true,
			errorMsg:    "item title is required",
		},
		{
			name: "empty Guid",
			item: xmlItem{
				Title:     "Valid Episode",
				Guid:      "",
				Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
			},
			expectError: true,
			errorMsg:    "item guid is required",
		},
		{
			name: "empty enclosure URL",
			item: xmlItem{
				Title:     "Valid Episode",
				Guid:      "valid-guid",
				Enclosure: xmlEnclosure{URL: "", Length: 1024, Type: Mp3},
			},
			expectError: true,
			errorMsg:    "item enclosure url is required",
		},
		{
			name: "zero enclosure length",
			item: xmlItem{
				Title:     "Valid Episode",
				Guid:      "valid-guid",
				Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 0, Type: Mp3},
			},
			expectError: true,
			errorMsg:    "item enclosure length must be greater than zero",
		},
		{
			name: "negative enclosure length",
			item: xmlItem{
				Title:     "Valid Episode",
				Guid:      "valid-guid",
				Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: -1, Type: Mp3},
			},
			expectError: true,
			errorMsg:    "item enclosure length must be greater than zero",
		},
		{
			name: "empty enclosure type",
			item: xmlItem{
				Title:     "Valid Episode",
				Guid:      "valid-guid",
				Enclosure: xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: ""},
			},
			expectError: true,
			errorMsg:    "item enclosure type is required",
		},
		{
			name: "valid explicit true",
			item: xmlItem{
				Title:          "Valid Episode",
				Guid:           "valid-guid",
				Enclosure:      xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
				ItunesExplicit: ExplicitTrue,
			},
			expectError: false,
		},
		{
			name: "valid explicit false",
			item: xmlItem{
				Title:          "Valid Episode",
				Guid:           "valid-guid",
				Enclosure:      xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
				ItunesExplicit: ExplicitFalse,
			},
			expectError: false,
		},
		{
			name: "empty explicit (valid)",
			item: xmlItem{
				Title:          "Valid Episode",
				Guid:           "valid-guid",
				Enclosure:      xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
				ItunesExplicit: "",
			},
			expectError: false,
		},
		{
			name: "invalid explicit value",
			item: xmlItem{
				Title:          "Valid Episode",
				Guid:           "valid-guid",
				Enclosure:      xmlEnclosure{URL: "https://example.com/audio.mp3", Length: 1024, Type: Mp3},
				ItunesExplicit: "invalid",
			},
			expectError: true,
			errorMsg:    "item itunes:explicit must be either 'true' or 'false'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.validate()

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

func TestXmlCDATA(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "simple text",
			data:     "Simple description",
			expected: "<![CDATA[Simple description]]>",
		},
		{
			name:     "text with HTML tags",
			data:     "Description with <b>bold</b> text",
			expected: "<![CDATA[Description with <b>bold</b> text]]>",
		},
		{
			name:     "text with special characters",
			data:     "Description with & ampersands and < > brackets",
			expected: "<![CDATA[Description with & ampersands and < > brackets]]>",
		},
		{
			name:     "empty text",
			data:     "",
			expected: "", // Empty CDATA sections are omitted entirely
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cdata := xmlCDATA{Data: tt.data}

			// Marshal to XML to test CDATA handling
			xmlData, err := xml.Marshal(cdata)
			if err != nil {
				t.Fatalf("Failed to marshal CDATA: %v", err)
			}

			xmlString := string(xmlData)
			if !strings.Contains(xmlString, tt.expected) {
				t.Errorf("Expected XML to contain '%s', got '%s'", tt.expected, xmlString)
			}
		})
	}
}

func TestXmlCategoryStructure(t *testing.T) {
	// Test simple category
	category := xmlItunesCategory{
		Text: "Technology",
	}

	xmlData, err := xml.Marshal(category)
	if err != nil {
		t.Fatalf("Failed to marshal category: %v", err)
	}

	xmlString := string(xmlData)
	if !strings.Contains(xmlString, `text="Technology"`) {
		t.Errorf("Expected category XML to contain text attribute, got: %s", xmlString)
	}

	// Test category with subcategories
	categoryWithSub := xmlItunesCategory{
		Text: "Society & Culture",
		ItunesCategories: []xmlItunesCategory{
			{Text: "Documentary"},
		},
	}

	xmlData, err = xml.Marshal(categoryWithSub)
	if err != nil {
		t.Fatalf("Failed to marshal category with subcategories: %v", err)
	}

	xmlString = string(xmlData)
	if !strings.Contains(xmlString, `text="Society &amp; Culture"`) {
		t.Errorf("Expected category XML to contain escaped ampersand, got: %s", xmlString)
	}

	if !strings.Contains(xmlString, `text="Documentary"`) {
		t.Errorf("Expected subcategory XML to be present, got: %s", xmlString)
	}
}

func TestXmlEnclosureAttributes(t *testing.T) {
	enclosure := xmlEnclosure{
		URL:    "https://example.com/episode.mp3",
		Length: 1024000,
		Type:   Mp3,
	}

	xmlData, err := xml.Marshal(enclosure)
	if err != nil {
		t.Fatalf("Failed to marshal enclosure: %v", err)
	}

	xmlString := string(xmlData)

	expectedAttributes := []string{
		`url="https://example.com/episode.mp3"`,
		`length="1024000"`,
		`type="audio/mpeg"`,
	}

	for _, attr := range expectedAttributes {
		if !strings.Contains(xmlString, attr) {
			t.Errorf("Expected enclosure XML to contain '%s', got: %s", attr, xmlString)
		}
	}
}

func TestXmlOwnerStructure(t *testing.T) {
	owner := xmlItunesOwner{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	xmlData, err := xml.Marshal(owner)
	if err != nil {
		t.Fatalf("Failed to marshal owner: %v", err)
	}

	xmlString := string(xmlData)

	expectedElements := []string{
		"<itunes:name>John Doe</itunes:name>",
		"<itunes:email>john@example.com</itunes:email>",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(xmlString, elem) {
			t.Errorf("Expected owner XML to contain '%s', got: %s", elem, xmlString)
		}
	}
}

func TestXmlTranscriptAttributes(t *testing.T) {
	transcript := xmlPodcastTranscript{
		Url:  "https://example.com/transcript.vtt",
		Type: string(TranscriptVtt),
	}

	xmlData, err := xml.Marshal(transcript)
	if err != nil {
		t.Fatalf("Failed to marshal transcript: %v", err)
	}

	xmlString := string(xmlData)

	expectedAttributes := []string{
		`url="https://example.com/transcript.vtt"`,
		`type="text/vtt"`,
	}

	for _, attr := range expectedAttributes {
		if !strings.Contains(xmlString, attr) {
			t.Errorf("Expected transcript XML to contain '%s', got: %s", attr, xmlString)
		}
	}
}

//goland:noinspection GoBoolExpressions
func TestConstants(t *testing.T) {
	// Test that constants match Apple Podcast requirements
	if xmlRssVersion != "2.0" {
		t.Errorf("RSS version should be '2.0', got '%s'", xmlRssVersion)
	}

	if xmlItunesNS != "http://www.itunes.com/dtds/podcast-1.0.dtd" {
		t.Errorf("iTunes namespace incorrect: %s", xmlItunesNS)
	}

	if xmlContentNS != "http://purl.org/rss/1.0/modules/content/" {
		t.Errorf("Content namespace incorrect: %s", xmlContentNS)
	}

	if xmlPodcastNS != "https://podcastindex.org/namespace/1.0" {
		t.Errorf("Podcast namespace incorrect: %s", xmlPodcastNS)
	}
}
