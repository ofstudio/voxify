package feedcast

// Category represents a podcast category with optional subcategories.
//
// Select the category that best reflects the content of your show.
// If available, you can also define a subcategory.
// Although you can specify more than one category and subcategory in your RSS feed,
// Apple Podcasts only recognizes the first category and subcategory.
//
// When specifying categories and subcategories, be sure to properly escape ampersands. For example:
//
// Single category:
//
//	<itunes:category text="History" />
//
// Category with ampersand:
//
//	<itunes:category text="Kids &amp; Family" />
//
// Category with subcategory:
//
//	<itunes:category text="Society &amp; Culture">
//	    <itunes:category text="Documentary" />
//	</itunes:category>
//
// Multiple categories:
//
//	<itunes:category text="Society &amp; Culture">
//	    <itunes:category text="Documentary" />
//	</itunes:category>
//	<itunes:category text="Health">
//	    <itunes:category text="Mental Health" />
//	</itunes:category>
//
// As of September 2025, the valid categories and subcategories are:
//
//	Arts
//	   Books
//	   Design
//	   Fashion & Beauty
//	   Food
//	   Performing Arts
//	   Visual Arts
//	Business
//	   Careers
//	   Entrepreneurship
//	   Investing
//	   Management
//	   Marketing
//	   Non-Profit
//	Comedy
//	   Comedy Interviews
//	   Improv
//	   Stand-Up
//	Education
//	   Courses
//	   How To
//	   Language Learning
//	   Self-Improvement
//	Fiction
//	   Comedy Fiction
//	   Drama
//	   Science Fiction
//	Government
//	History
//	Health & Fitness
//	   Alternative Health
//	   Fitness
//	   Medicine
//	   Mental Health
//	   Nutrition
//	   Sexuality
//	Kids & Family
//	   Education for Kids
//	   Parenting
//	   Pets & Animals
//	   Stories for Kids
//	Leisure
//	   Animation & Manga
//	   Automotive
//	   Aviation
//	   Crafts
//	   Games
//	   Hobbies
//	   Home & Garden
//	   Video Games
//	Music
//	   Music Commentary
//	   Music History
//	   Music Interviews
//	News
//	   Business News
//	   Daily News
//	   Entertainment News
//	   News Commentary
//	   Politics
//	   Sports News
//	   Tech News
//	Religion & Spirituality
//	   Buddhism
//	   Christianity
//	   Hinduism
//	   Islam
//	   Judaism
//	   Religion
//	   Spirituality
//	Science
//	   Astronomy
//	   Chemistry
//	   Earth Sciences
//	   Life Sciences
//	   Mathematics
//	   Natural Sciences
//	   Nature
//	   Physics
//	   Social Sciences
//	Society & Culture
//	   Documentary
//	   Personal Journals
//	   Philosophy
//	   Places & Travel
//	   Relationships
//	Sports
//	   Baseball
//	   Basketball
//	   Cricket
//	   Fantasy Sports
//	   Football
//	   Golf
//	   Hockey
//	   Rugby
//	   Running
//	   Soccer
//	   Swimming
//	   Tennis
//	   Volleyball
//	   Wilderness
//	   Wrestling
//	Technology
//	True Crime
//	TV & Film
//	   After Shows
//	   Film History
//	   Film Interviews
//	   Film Reviews
//	   TV Reviews
//
// See Apple Podcast categories:
// https://podcasters.apple.com/support/1691-apple-podcasts-categories
type Category struct {
	Text          string
	Subcategories []string
}

// NewCategory creates a new Category with optional subcategories.
func NewCategory(text string, subcategories ...string) Category {
	return Category{
		Text:          text,
		Subcategories: subcategories,
	}
}
