package templates

import (
	"html/template"
	"strings"
	"time"

	"github.com/ofstudio/voxify/internal/domain"
)

// tmplFns - template custom functions
var tmplFns = template.FuncMap{
	"truncate":       tmplTruncate,
	"categoriesList": tmplCategoriesList,
	"episodeDate":    tmplEpisodeDate,
	"episodeURL":     tmplEpisodeURL,
}

// tmplTruncate - template helper to truncate a string to n runes.
func tmplTruncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return string(r)
	}
	return string(r[:n-1]) + "…"
}

// tmplCategoriesList - template helper to flatten categories and subcategories into a single string slice.
func tmplCategoriesList(cats []domain.FeedCategory) []string {
	var result []string
	for _, c := range cats {
		result = append(result, c.Text)
		if len(c.Subcategories) > 0 {
			for _, sc := range c.Subcategories {
				result = append(result, sc)
			}
		}
	}
	return result
}

// tmplEpisodeDate - template helper to format episode date.
func tmplEpisodeDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("Jan 02, 2006")
}

// tmplEpisodeURL - template helper to get the best available URL for an episode.
func tmplEpisodeURL(e domain.Episode) string {
	if u := strings.TrimSpace(e.CanonicalURL); u != "" {
		return u
	}
	if u := strings.TrimSpace(e.OriginalURL); u != "" {
		return u
	}
	return "#"
}
