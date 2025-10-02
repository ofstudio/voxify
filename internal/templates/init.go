package templates

import (
	"context"
	"fmt"
	"html/template"
)

func Init(_ context.Context) error {
	var err error

	// Parse landing page template
	LandingTemplate, err = template.New("landing").Funcs(tmplFns).Parse(landingTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse landing template: %w", err)
	}

	// Parse /info response template
	FeedInfoTemplate, err = template.New("feed_info").Funcs(tmplFns).Parse(feedInfoTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse feed info template: %w", err)
	}

	return nil
}
