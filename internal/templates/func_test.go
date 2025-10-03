package templates

import (
	"bytes"
	"html/template"
	"testing"
	"time"

	"github.com/ofstudio/voxify/internal/domain"
	"github.com/stretchr/testify/suite"
)

// TestTemplateFunctionsSuite defines the test suite for template functions
type TestTemplateFunctionsSuite struct {
	suite.Suite
}

// TestTruncateFunction tests the tmplTruncate function
func (suite *TestTemplateFunctionsSuite) TestTruncateFunction() {
	suite.Run("basic truncation", func() {
		result := tmplTruncate("Hello World", 6)
		suite.Equal("Hello…", result, "Should truncate and add ellipsis")
	})

	suite.Run("no truncation needed", func() {
		result := tmplTruncate("Hello", 10)
		suite.Equal("Hello", result, "Should return original string when no truncation needed")
	})

	suite.Run("exact length", func() {
		result := tmplTruncate("Hello", 5)
		suite.Equal("Hello", result, "Should return original string when exact length")
	})

	suite.Run("zero or negative length", func() {
		suite.Equal("", tmplTruncate("Hello", 0), "Should return empty string for zero length")
		suite.Equal("", tmplTruncate("Hello", -1), "Should return empty string for negative length")
	})

	suite.Run("with whitespace", func() {
		result := tmplTruncate("  Hello World  ", 6)
		suite.Equal("Hello…", result, "Should trim whitespace before truncating")
	})

	suite.Run("unicode characters", func() {
		result := tmplTruncate("Привет мир", 7)
		suite.Equal("Привет…", result, "Should handle unicode characters correctly")
	})

	suite.Run("empty string", func() {
		result := tmplTruncate("", 5)
		suite.Equal("", result, "Should handle empty string")
	})

}

// TestCategoriesListFunction tests the tmplCategoriesList function
func (suite *TestTemplateFunctionsSuite) TestCategoriesListFunction() {
	suite.Run("empty categories", func() {
		result := tmplCategoriesList([]domain.FeedCategory{})
		suite.Empty(result, "Should return empty slice for empty input")
	})

	suite.Run("categories without subcategories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology"},
			{Text: "Education"},
		}
		result := tmplCategoriesList(categories)
		expected := []string{"Technology", "Education"}
		suite.Equal(expected, result)
	})

	suite.Run("categories with subcategories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology", Subcategories: []string{"Software", "Hardware"}},
			{Text: "Education", Subcategories: []string{"Online", "Courses"}},
		}
		result := tmplCategoriesList(categories)
		expected := []string{"Technology", "Software", "Hardware", "Education", "Online", "Courses"}
		suite.Equal(expected, result)
	})

	suite.Run("mixed categories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology", Subcategories: []string{"Software"}},
			{Text: "Education"}, // no subcategories
			{Text: "Business", Subcategories: []string{"Marketing", "Finance"}},
		}
		result := tmplCategoriesList(categories)
		expected := []string{"Technology", "Software", "Education", "Business", "Marketing", "Finance"}
		suite.Equal(expected, result)
	})

	suite.Run("nil input", func() {
		result := tmplCategoriesList(nil)
		suite.Empty(result, "Should handle nil input gracefully")
	})

}

// TestCategoriesEnumFunction tests the tmplCategoriesEnum function
func (suite *TestTemplateFunctionsSuite) TestCategoriesEnumFunction() {
	suite.Run("empty categories", func() {
		result := tmplCategoriesEnum([]domain.FeedCategory{})
		suite.Equal("", result, "Should return empty string for empty input")
	})

	suite.Run("single category without subcategories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology"},
		}
		result := tmplCategoriesEnum(categories)
		suite.Equal("Technology", result)
	})

	suite.Run("multiple categories without subcategories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology"},
			{Text: "Education"},
			{Text: "Business"},
		}
		result := tmplCategoriesEnum(categories)
		suite.Equal("Technology, Education, Business", result)
	})

	suite.Run("categories with subcategories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology", Subcategories: []string{"Software", "Hardware"}},
			{Text: "Education", Subcategories: []string{"Online"}},
		}
		result := tmplCategoriesEnum(categories)
		suite.Equal("Technology, Software, Hardware, Education, Online", result)
	})

	suite.Run("mixed categories", func() {
		categories := []domain.FeedCategory{
			{Text: "Technology", Subcategories: []string{"Software"}},
			{Text: "Education"}, // no subcategories
			{Text: "Business", Subcategories: []string{"Marketing", "Finance"}},
		}
		result := tmplCategoriesEnum(categories)
		suite.Equal("Technology, Software, Education, Business, Marketing, Finance", result)
	})

	suite.Run("nil input", func() {
		result := tmplCategoriesEnum(nil)
		suite.Equal("", result, "Should handle nil input gracefully")
	})

	suite.Run("categories with special characters", func() {
		categories := []domain.FeedCategory{
			{Text: "Arts & Crafts"},
			{Text: "Health & Fitness"},
		}
		result := tmplCategoriesEnum(categories)
		suite.Equal("Arts & Crafts, Health & Fitness", result)
	})

	suite.Run("in template context", func() {
		tmpl, err := template.New("test").Funcs(tmplFns).Parse(`{{ categoriesEnum . }}`)
		suite.Require().NoError(err)

		categories := []domain.FeedCategory{
			{Text: "Technology"},
			{Text: "Education"},
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, categories)
		suite.Require().NoError(err)

		suite.Equal("Technology, Education", buf.String())
	})
}

// TestEpisodeDateFunction tests the tmplEpisodeDate function
func (suite *TestTemplateFunctionsSuite) TestEpisodeDateFunction() {
	suite.Run("valid date", func() {
		testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)
		result := tmplEpisodeDate(testTime)
		suite.Equal("Dec 25, 2023", result)
	})

	suite.Run("zero time", func() {
		result := tmplEpisodeDate(time.Time{})
		suite.Equal("", result, "Should return empty string for zero time")
	})
}

// TestEpisodeURLFunction tests the tmplEpisodeURL function
func (suite *TestTemplateFunctionsSuite) TestEpisodeURLFunction() {
	suite.Run("canonical URL present", func() {
		episode := domain.Episode{
			CanonicalURL: "https://example.com/canonical",
			OriginalURL:  "https://example.com/original",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("https://example.com/canonical", result, "Should prefer canonical URL")
	})

	suite.Run("only original URL present", func() {
		episode := domain.Episode{
			OriginalURL: "https://example.com/original",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("https://example.com/original", result, "Should use original URL when canonical is empty")
	})

	suite.Run("canonical URL with whitespace", func() {
		episode := domain.Episode{
			CanonicalURL: "  https://example.com/canonical  ",
			OriginalURL:  "https://example.com/original",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("https://example.com/canonical", result, "Should trim whitespace from canonical URL")
	})

	suite.Run("original URL with whitespace", func() {
		episode := domain.Episode{
			OriginalURL: "  https://example.com/original  ",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("https://example.com/original", result, "Should trim whitespace from original URL")
	})

	suite.Run("no URLs present", func() {
		episode := domain.Episode{}
		result := tmplEpisodeURL(episode)
		suite.Equal("#", result, "Should return # when no URLs available")
	})

	suite.Run("empty URLs", func() {
		episode := domain.Episode{
			CanonicalURL: "",
			OriginalURL:  "",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("#", result, "Should return # when URLs are empty")
	})

	suite.Run("whitespace only URLs", func() {
		episode := domain.Episode{
			CanonicalURL: "   ",
			OriginalURL:  "   ",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("#", result, "Should return # when URLs are whitespace only")
	})

	suite.Run("special characters in URL", func() {
		episode := domain.Episode{
			CanonicalURL: "https://example.com/видео?query=тест&param=значение",
		}
		result := tmplEpisodeURL(episode)
		suite.Equal("https://example.com/видео?query=тест&param=значение", result)
	})

	suite.Run("in episode card template", func() {
		tmpl, err := template.New("test").Funcs(tmplFns).Parse(`
			<a class="card" href="{{ episodeURL . }}" target="_blank" rel="noopener">
				<h3>{{ .Title }}</h3>
			</a>
		`)
		suite.Require().NoError(err)

		episode := domain.Episode{
			Title:        "Test Episode",
			CanonicalURL: "https://youtube.com/watch?v=123",
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, episode)
		suite.Require().NoError(err)

		html := buf.String()
		suite.Contains(html, `href="https://youtube.com/watch?v=123"`)
		suite.Contains(html, "Test Episode")
	})

	suite.Run("fallback to hash in template", func() {
		tmpl, err := template.New("test").Funcs(tmplFns).Parse(`<a href="{{ episodeURL . }}">Link</a>`)
		suite.Require().NoError(err)

		episode := domain.Episode{
			Title: "Test Episode",
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, episode)
		suite.Require().NoError(err)

		suite.Equal(`<a href="#">Link</a>`, buf.String())
	})
}

// TestTemplateFunctions runs the test suite
func TestTemplateFunctions(t *testing.T) {
	suite.Run(t, new(TestTemplateFunctionsSuite))
}
