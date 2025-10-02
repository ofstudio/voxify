package templates

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/stretchr/testify/suite"
)

// TestFeedInfoTemplateSuite defines the test suite for feed-info template
type TestFeedInfoTemplateSuite struct {
	suite.Suite
	template *template.Template
}

// SetupSuite parses the feed info template once for all tests
func (suite *TestFeedInfoTemplateSuite) SetupSuite() {
	var err error
	suite.template, err = template.New("feed-info").Funcs(tmplFns).Parse(feedInfoTemplateHTML)
	suite.Require().NoError(err, "Failed to parse feed-info template")
}

// TestFeedInfoTemplateGeneration basic rendering
func (suite *TestFeedInfoTemplateSuite) TestFeedInfoTemplateGeneration() {
	data := domain.FeedInfo{
		Title:        "Test Podcast",
		Description:  "A test podcast for feed-info template",
		Language:     "en",
		RSSLink:      "https://example.com/rss.xml",
		EpisodeCount: 5, // Set episode count > 0 to ensure RSS is displayed
	}

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)

	suite.NoError(err, "Template execution should not fail")
	suite.NotEmpty(buf.String(), "Generated text should not be empty")

	html := buf.String()
	suite.Contains(html, "Test Podcast")
	suite.Contains(html, "A test podcast for feed-info template")
	suite.Contains(html, "Language: en")
	suite.Contains(html, "📡 RSS: https://example.com/rss.xml")
	suite.Contains(html, "🎧 Number of episodes: 5")
}

// TestFeedInfoTemplateWithEmptyFields checks optional parts are omitted
func (suite *TestFeedInfoTemplateSuite) TestFeedInfoTemplateWithEmptyFields() {
	data := domain.FeedInfo{
		Title:        "Minimal Podcast",
		EpisodeCount: 0, // Explicitly set to 0 to test RSS hiding logic
	}

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)
	suite.Require().NoError(err)

	html := buf.String()

	// No author, no keywords, no image/website, no episodes
	suite.NotContains(html, "By <strong>")
	suite.Contains(html, "📭 No episodes yet")
	suite.NotContains(html, "🔑 Keywords:")
	suite.NotContains(html, "Artwork")
	suite.NotContains(html, "Website")
	suite.NotContains(html, "🔞 Explicit content")

	// When no episodes exist, RSS elements should not be present
	suite.NotContains(html, "📡 RSS:")
	suite.NotContains(html, "RSS feed")
}

// TestFeedInfoTemplateWithFullFields ensures all sections render correctly
func (suite *TestFeedInfoTemplateSuite) TestFeedInfoTemplateWithFullFields() {
	data := domain.FeedInfo{
		Title:       "Complete Podcast",
		Description: "A fully populated podcast info",
		Language:    "ru",
		Categories: []domain.FeedCategory{
			{Text: "Technology", Subcategories: []string{"Software", "Hardware"}},
		},
		Keywords:     "go, podcast",
		Author:       "Podcast Author",
		ImageUrl:     "https://example.com/cover.jpg",
		WebsiteLink:  "https://example.com",
		EpisodeCount: 3,
		Explicit:     true,
		RSSLink:      "https://example.com/rss.xml",
	}

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)
	suite.Require().NoError(err)

	html := buf.String()

	// Basic fields
	suite.Contains(html, "Complete Podcast")
	suite.Contains(html, "A fully populated podcast info")
	suite.Contains(html, "Language: ru")

	// Optional fields
	suite.Contains(html, "Podcast Author")
	suite.Contains(html, "🔑 Keywords: go, podcast")
	suite.Contains(html, "<a href=\"https://example.com/cover.jpg\">Artwork</a>")
	suite.Contains(html, "<a href=\"https://example.com\">Website</a>")
	suite.Contains(html, "🎧 Number of episodes: 3")
	suite.Contains(html, "🔞 Explicit content")

	// Categories are flattened by categoriesList; ensure main and subcategory appear
	suite.Contains(html, "Technology")
	suite.Contains(html, "Software")

	// When episodes exist, RSS elements should be present
	suite.Contains(html, "📡 RSS: https://example.com/rss.xml")
}

// TestFeedInfoTemplate runs the suite
func TestFeedInfoTemplate(t *testing.T) {
	suite.Run(t, new(TestFeedInfoTemplateSuite))
}
