package feedcast

import (
	"testing"
)

func TestNewCategory(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		subcategories  []string
		expectedText   string
		expectedSubLen int
	}{
		{
			name:           "simple category",
			text:           "Technology",
			subcategories:  nil,
			expectedText:   "Technology",
			expectedSubLen: 0,
		},
		{
			name:           "category with one subcategory",
			text:           "Business",
			subcategories:  []string{"Entrepreneurship"},
			expectedText:   "Business",
			expectedSubLen: 1,
		},
		{
			name:           "category with multiple subcategories",
			text:           "Arts",
			subcategories:  []string{"Books", "Design", "Fashion & Beauty"},
			expectedText:   "Arts",
			expectedSubLen: 3,
		},
		{
			name:           "category with ampersand",
			text:           "Kids & Family",
			subcategories:  []string{"Education for Kids"},
			expectedText:   "Kids & Family",
			expectedSubLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := NewCategory(tt.text, tt.subcategories...)

			if category.Text != tt.expectedText {
				t.Errorf("Expected text %s, got %s", tt.expectedText, category.Text)
			}

			if len(category.Subcategories) != tt.expectedSubLen {
				t.Errorf("Expected %d subcategories, got %d", tt.expectedSubLen, len(category.Subcategories))
			}

			for i, sub := range tt.subcategories {
				if i < len(category.Subcategories) && category.Subcategories[i] != sub {
					t.Errorf("Expected subcategory %s, got %s", sub, category.Subcategories[i])
				}
			}
		})
	}
}

// TestApplePodcastCategories tests that we can create categories that match
// the official Apple Podcasts category list as of September 2025
func TestApplePodcastCategories(t *testing.T) {
	// Test main categories without subcategories
	mainOnlyCategories := []string{
		"Comedy",
		"Education",
		"Fiction",
		"Government",
		"History",
		"Music",
		"Science",
		"True Crime",
	}

	for _, cat := range mainOnlyCategories {
		t.Run("main_only_"+cat, func(t *testing.T) {
			category := NewCategory(cat)
			if category.Text != cat {
				t.Errorf("Expected category text %s, got %s", cat, category.Text)
			}
			if len(category.Subcategories) != 0 {
				t.Errorf("Expected no subcategories for %s, got %d", cat, len(category.Subcategories))
			}
		})
	}

	// Test categories with subcategories (sample)
	categoriesWithSubs := map[string][]string{
		"Arts": {
			"Books",
			"Design",
			"Fashion & Beauty",
			"Food",
			"Performing Arts",
			"Visual Arts",
		},
		"Business": {
			"Careers",
			"Entrepreneurship",
			"Investing",
			"Management",
			"Marketing",
			"Non-Profit",
		},
		"Health & Fitness": {
			"Alternative Health",
			"Fitness",
			"Medicine",
			"Mental Health",
			"Nutrition",
			"Sexuality",
		},
		"Kids & Family": {
			"Education for Kids",
			"Parenting",
			"Pets & Animals",
			"Stories for Kids",
		},
		"Society & Culture": {
			"Documentary",
			"Personal Journals",
			"Philosophy",
			"Places & Travel",
			"Relationships",
		},
	}

	for mainCat, subs := range categoriesWithSubs {
		t.Run("with_subs_"+mainCat, func(t *testing.T) {
			category := NewCategory(mainCat, subs...)

			if category.Text != mainCat {
				t.Errorf("Expected category text %s, got %s", mainCat, category.Text)
			}

			if len(category.Subcategories) != len(subs) {
				t.Errorf("Expected %d subcategories for %s, got %d", len(subs), mainCat, len(category.Subcategories))
			}

			for i, expectedSub := range subs {
				if i < len(category.Subcategories) && category.Subcategories[i] != expectedSub {
					t.Errorf("Expected subcategory %s, got %s", expectedSub, category.Subcategories[i])
				}
			}
		})
	}
}

// TestCategorySpecialCharacters tests categories with special characters
// that need proper XML escaping
func TestCategorySpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		category string
		expected string
	}{
		{
			name:     "ampersand in main category",
			category: "Kids & Family",
			expected: "Kids & Family", // Raw text, XML escaping happens during marshaling
		},
		{
			name:     "ampersand in subcategory",
			category: "Fashion & Beauty",
			expected: "Fashion & Beauty",
		},
		{
			name:     "category without special chars",
			category: "Technology",
			expected: "Technology",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category := NewCategory(tt.category)
			if category.Text != tt.expected {
				t.Errorf("Expected category text %s, got %s", tt.expected, category.Text)
			}
		})
	}
}

// TestCategoryValidationInFeed tests that categories work correctly
// when integrated into a feed
func TestCategoryValidationInFeed(t *testing.T) {
	tests := []struct {
		name       string
		categories []Category
		shouldPass bool
	}{
		{
			name: "single valid category",
			categories: []Category{
				NewCategory("Technology"),
			},
			shouldPass: true,
		},
		{
			name: "multiple valid categories",
			categories: []Category{
				NewCategory("Technology"),
				NewCategory("Business", "Entrepreneurship"),
			},
			shouldPass: true,
		},
		{
			name: "category with multiple subcategories",
			categories: []Category{
				NewCategory("Arts", "Books", "Design", "Food"),
			},
			shouldPass: true,
		},
		{
			name: "category with special characters",
			categories: []Category{
				NewCategory("Kids & Family", "Education for Kids"),
			},
			shouldPass: true,
		},
		{
			name:       "no categories",
			categories: []Category{},
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channelData := FeedData{
				Title:       "Test Podcast",
				Description: "Test description",
				Image:       "https://example.com/image.jpg",
				Language:    "en",
				Explicit:    ExplicitFalse,
				Categories:  tt.categories,
			}

			feed := NewFeed(channelData)

			// Add a valid item since it's required for validation
			if tt.shouldPass {
				itemData := ItemData{
					Title: "Test Episode",
					Guid:  "test-guid",
					Enclosure: Enclosure{
						URL:    "https://example.com/audio.mp3",
						Length: 1024,
						Type:   Mp3,
					},
				}
				item := NewItem(itemData)
				feed.AddItem(item)
			}

			err := feed.Validate()

			if tt.shouldPass && err != nil {
				t.Errorf("Expected validation to pass but got error: %v", err)
			} else if !tt.shouldPass && err == nil {
				t.Error("Expected validation to fail but it passed")
			}
		})
	}
}

// TestCategoryDocumentationCompliance ensures our categories match
// the documented Apple Podcasts requirements
func TestCategoryDocumentationCompliance(t *testing.T) {
	// Test that we can create all the documented categories
	// This serves as documentation and ensures we support the full spec

	// Sample of main categories from Apple's documentation
	documentedCategories := []string{
		"Arts",
		"Business",
		"Comedy",
		"Education",
		"Fiction",
		"Government",
		"Health & Fitness",
		"History",
		"Kids & Family",
		"Leisure",
		"Music",
		"News",
		"Religion & Spirituality",
		"Science",
		"Society & Culture",
		"Sports",
		"Technology",
		"True Crime",
		"TV & Film",
	}

	for _, cat := range documentedCategories {
		t.Run("documented_category_"+cat, func(t *testing.T) {
			category := NewCategory(cat)
			if category.Text != cat {
				t.Errorf("Failed to create documented category %s", cat)
			}
		})
	}

	// Test sample subcategories for verification
	subcategoryTests := map[string][]string{
		"Arts":             {"Books", "Design", "Fashion & Beauty", "Food", "Performing Arts", "Visual Arts"},
		"Business":         {"Careers", "Entrepreneurship", "Investing", "Management", "Marketing", "Non-Profit"},
		"Comedy":           {"Comedy Interviews", "Improv", "Stand-Up"},
		"Education":        {"Courses", "How To", "Language Learning", "Self-Improvement"},
		"Fiction":          {"Comedy Fiction", "Drama", "Science Fiction"},
		"Health & Fitness": {"Alternative Health", "Fitness", "Medicine", "Mental Health", "Nutrition", "Sexuality"},
		"Kids & Family":    {"Education for Kids", "Parenting", "Pets & Animals", "Stories for Kids"},
		"News":             {"Business News", "Daily News", "Entertainment News", "News Commentary", "Politics", "Sports News", "Tech News"},
		"Sports":           {"Baseball", "Basketball", "Cricket", "Fantasy Sports", "Football", "Golf", "Hockey", "Rugby", "Running", "Soccer", "Swimming", "Tennis", "Volleyball", "Wilderness", "Wrestling"},
		"TV & Film":        {"After Shows", "Film History", "Film Interviews", "Film Reviews", "TV Reviews"},
	}

	for mainCat, subs := range subcategoryTests {
		t.Run("subcategories_"+mainCat, func(t *testing.T) {
			for _, sub := range subs {
				category := NewCategory(mainCat, sub)

				if len(category.Subcategories) == 0 {
					t.Errorf("Failed to add subcategory %s to %s", sub, mainCat)
				}

				found := false
				for _, actualSub := range category.Subcategories {
					if actualSub == sub {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("Subcategory %s not found in %s", sub, mainCat)
				}
			}
		})
	}
}
