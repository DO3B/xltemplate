package templateengine

import (
	"bytes"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type TemplateEngine struct {
	TemplateName string
	Variables    map[string]interface{}
	Source       string
	Patterns     []string
}

func NewTemplateEngine(
	templateName string, variables map[string]interface{}, source string, patterns []string) *TemplateEngine {
	return &TemplateEngine{
		TemplateName: templateName,
		Variables:    variables,
		Source:       source,
		Patterns:     patterns,
	}
}

func (templateEngine *TemplateEngine) Parse() (string, error) {
	var err error

	var tpl template.Template
	slog.Debug("Loading variables from file", "variables", templateEngine.Variables)
	slog.Debug("Loading source file", "source", templateEngine.Source)
	slog.Debug("Loading patterns", "patterns", templateEngine.Patterns)
	// Add custom include and sprig lib functions to the template
	var funcMap template.FuncMap = map[string]interface{}{}
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		return buf.String(), nil
	}

	tpl.Funcs(sprig.TxtFuncMap()).Funcs(funcMap)

	// Create the main template from the source
	tpl.New(templateEngine.TemplateName).Parse(templateEngine.Source)

	// Add patterns to template
	_, err = tpl.ParseFiles(templateEngine.Patterns...)
	if err != nil {
		return "", err
	}

	result := bytes.NewBuffer(nil)
	err = tpl.ExecuteTemplate(result, templateEngine.TemplateName, templateEngine.Variables)
	if err != nil {
		panic(err)
	}

	return result.String(), nil
}
