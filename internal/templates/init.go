package templates

import (
	"context"
	"fmt"
	"html/template"
)

func Init(_ context.Context) error {
	var err error
	LandingTemplate, err = template.New("landing").Funcs(tmplFns).Parse(landingTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse landing template: %w", err)
	}
	return nil
}
