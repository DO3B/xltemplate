package templateengine

import (
	"bytes"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func Parse(templateName string, variables map[string]interface{}, source string, patterns []string) (string, error) {
	var err error

	var tpl template.Template
	slog.Debug("Loading variables from file", "variables", variables)
	slog.Debug("Loading source file", "source", source)
	slog.Debug("Loading patterns", "patterns", patterns)
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
	tpl.New(templateName).Parse(source)

	// Add patterns to template
	_, err = tpl.ParseFiles(patterns...)
	if err != nil {
		return "", err
	}

	result := bytes.NewBuffer(nil)
	err = tpl.ExecuteTemplate(result, templateName, variables)
	if err != nil {
		panic(err)
	}
	slog.Debug("Result", "result", result.String())

	return result.String(), nil
}
