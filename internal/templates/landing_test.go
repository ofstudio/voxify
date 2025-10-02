package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"
	"time"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/stretchr/testify/suite"
)

// TestLandingTemplateSuite defines the test suite for landing template
type TestLandingTemplateSuite struct {
	suite.Suite
	template *template.Template
}

// SetupSuite runs once before all tests in the suite
func (suite *TestLandingTemplateSuite) SetupSuite() {
	var err error
	suite.template, err = template.New("landing").Funcs(tmplFns).Parse(landingTemplateHTML)
	suite.Require().NoError(err, "Failed to parse landing template")
}

// TestGeneration tests basic template generation without errors
func (suite *TestLandingTemplateSuite) TestGeneration() {
	data := suite.createTestData()

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)

	suite.NoError(err, "Template execution should not fail")
	suite.NotEmpty(buf.String(), "Generated HTML should not be empty")
}

// TestWithEmptyFeed tests template with minimal feed data
func (suite *TestLandingTemplateSuite) TestWithEmptyFeed() {
	data := LandingData{
		Feed:     domain.FeedInfo{Title: "Test Podcast"},
		Episodes: []*domain.Episode{},
	}

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)

	suite.NoError(err, "Template execution should not fail with empty feed")

	html := buf.String()
	suite.Contains(html, "Test Podcast", "Should contain feed title")
	suite.Contains(html, "No episodes yet", "Should show no episodes message")

	// When no episodes exist, RSS elements should not be present
	suite.NotContains(html, `<link rel="alternate" type="application/rss+xml"`, "Should not contain RSS link tag in head when no episodes")
	suite.NotContains(html, "RSS feed", "Should not show RSS feed button when no episodes")
}

// TestLandingTemplateWithFullFeed tests template with complete feed data
func (suite *TestLandingTemplateSuite) TestLandingTemplateWithFullFeed() {
	data := suite.createFullTestData()

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)

	suite.NoError(err, "Template execution should not fail with full feed")

	html := buf.String()

	// Test feed data rendering
	suite.Contains(html, data.Feed.Title, "Should contain feed title")
	suite.Contains(html, data.Feed.Description, "Should contain feed description")
	suite.Contains(html, data.Feed.Author, "Should contain feed author")
	suite.Contains(html, data.Feed.WebsiteLink, "Should contain website link")
	suite.Contains(html, data.Feed.RSSLink, "Should contain RSS link")
	suite.Contains(html, data.Feed.ImageUrl, "Should contain image URL")
	suite.Contains(html, data.Feed.Keywords, "Should contain keywords")

	// Test categories rendering
	suite.Contains(html, "Technology", "Should contain category")
	suite.Contains(html, "Software", "Should contain subcategory")

	// Test episodes rendering
	suite.Contains(html, "Episode 1", "Should contain episode title")
	suite.Contains(html, "Episode 2", "Should contain episode title")
	suite.NotContains(html, "No episodes yet", "Should not show no episodes message")
}

// TestManyEpisodes tests that only first 10 episodes are shown
func (suite *TestLandingTemplateSuite) TestManyEpisodes() {
	data := LandingData{
		Feed:     domain.FeedInfo{Title: "Test Podcast"},
		Episodes: make([]*domain.Episode, 15),
	}

	// Create 15 episodes
	for i := 0; i < 15; i++ {
		data.Episodes[i] = &domain.Episode{
			ID:          int64(i + 1),
			Title:       fmt.Sprintf("Episode %d", i+1),
			Description: fmt.Sprintf("Description for episode %d", i+1),
			Author:      "Test Author",
			CreatedAt:   time.Now().AddDate(0, 0, -i),
		}
	}

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, data)

	suite.NoError(err, "Template execution should not fail with many episodes")

	html := buf.String()

	// Should contain first 10 episodes
	for i := 1; i <= 10; i++ {
		suite.Contains(html, fmt.Sprintf("Episode %d", i), "Should contain episode %d", i)
	}

	// Should not contain episodes 11-15
	for i := 11; i <= 15; i++ {
		suite.NotContains(html, fmt.Sprintf("Episode %d", i), "Should not contain episode %d", i)
	}
}

// TestConditionalRendering tests conditional rendering logic
func (suite *TestLandingTemplateSuite) TestConditionalRendering() {
	// Test with minimal data
	minimalData := LandingData{
		Feed:     domain.FeedInfo{Title: "Minimal Podcast"},
		Episodes: []*domain.Episode{},
	}

	var buf bytes.Buffer
	err := suite.template.Execute(&buf, minimalData)
	suite.Require().NoError(err)

	html := buf.String()

	// Should not contain optional elements when data is missing
	suite.NotContains(html, "By <strong>", "Should not show author when empty")
	suite.NotContains(html, "Website", "Should not show website link when empty")
	suite.NotContains(html, "RSS feed", "Should not show RSS button when empty")
	suite.NotContains(html, "class=\"chip\"", "Should not show categories when empty")
	suite.NotContains(html, `<link rel="alternate" type="application/rss+xml"`, "Should not contain RSS link tag in head when no episodes")
}

// createTestData creates basic test data for template testing
func (suite *TestLandingTemplateSuite) createTestData() LandingData {
	return LandingData{
		Feed: domain.FeedInfo{
			Title:       "Test Podcast",
			Description: "A test podcast for template testing",
			Author:      "Test Author",
		},
		Episodes: []*domain.Episode{
			{
				ID:          1,
				Title:       "Episode 1",
				Description: "First test episode",
				Author:      "Test Author",
				CreatedAt:   time.Now().AddDate(0, 0, -1),
			},
		},
	}
}

// createFullTestData creates complete test data with all optional fields
func (suite *TestLandingTemplateSuite) createFullTestData() LandingData {
	return LandingData{
		Feed: domain.FeedInfo{
			Title:       "Complete Test Podcast",
			Description: "A complete test podcast with all fields",
			Author:      "Complete Author",
			WebsiteLink: "https://example.com/website",
			RSSLink:     "https://example.com/rss.xml",
			ImageUrl:    "https://example.com/cover.jpg",
			Keywords:    "test, podcast, example",
			Categories: []domain.FeedCategory{
				{Text: "Technology", Subcategories: []string{"Software", "Hardware"}},
				{Text: "Education", Subcategories: []string{}},
			},
		},
		Episodes: []*domain.Episode{
			{
				ID:            1,
				Title:         "Episode 1",
				Description:   "First complete test episode with a very long description that should be truncated when displayed in the template to ensure proper formatting and layout",
				ThumbnailFile: "thumb1.jpg",
				Author:        "Episode Author 1",
				CanonicalURL:  "https://example.com/episode1",
				CreatedAt:     time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
			},
			{
				ID:          2,
				Title:       "Episode 2",
				Description: "Second complete test episode",
				Author:      "Episode Author 2",
				OriginalURL: "https://example.com/original2",
				CreatedAt:   time.Date(2023, 12, 20, 15, 45, 0, 0, time.UTC),
			},
		},
	}
}

// TestLandingTemplateTestSuite runs the test suite
func TestLandingTemplate(t *testing.T) {
	suite.Run(t, new(TestLandingTemplateSuite))
}
