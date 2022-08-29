package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

type ProspectiveTemplate struct {
	Cia        string      `json:"cia" binding:"required"`
	Components []Component `json:"components" binding:"required"`
}

type Component struct {
	Type     string   `json:"type"`
	Packages []string `json:"packages"`
}

type Pack struct {
	Name  string
	AppId string
}

func main() {
	router := gin.Default()

	router.POST("/createTemplate", func(context *gin.Context) {
		var prospectiveTemplate ProspectiveTemplate
		if err := context.BindJSON(&prospectiveTemplate); err != nil {
			context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var variablesStringBuilder strings.Builder

		var variablesTpl template.Template
		templateName := filepath.Base("sample/variables.tmpl")
		// Create the template based on pattern name
		b, err := ioutil.ReadFile("sample/variables.tmpl")
		if err != nil {
			panic(err)
		}
		variablesTpl.New(templateName).Parse(string(b))
		//Applies the main template
		result := bytes.NewBuffer(nil)
		err = variablesTpl.ExecuteTemplate(result, templateName, prospectiveTemplate)
		if err != nil {
			panic(err)
		}

		variablesStringBuilder.WriteString(result.String() + "\r\n")

		for _, component := range prospectiveTemplate.Components {
			// Get pattern name based on a normalize version of component's name
			patternName := strings.ReplaceAll(strings.ToLower(component.Type), " ", "")

			var componentTpl template.Template
			// Get the pattern name
			templateName = filepath.Base("sample/components/" + patternName + ".tmpl")
			// Create the template based on pattern name
			b, err := ioutil.ReadFile("sample/components/" + patternName + ".tmpl")
			if err != nil {
				panic(err)
			}
			componentTpl.New(templateName).Parse(string(b))

			var packages []Pack
			for _, pack := range component.Packages {
				newPack := Pack{
					Name:  strings.Split(pack, "/")[len(strings.Split(pack, "/"))-1],
					AppId: pack,
				}
				packages = append(packages, newPack)
			}

			result = bytes.NewBuffer(nil)

			err = componentTpl.ExecuteTemplate(result, templateName, packages)
			if err != nil {
				context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			variablesStringBuilder.WriteString(result.String() + "\r\n")
		}
		fmt.Println(variablesStringBuilder.String())
		context.JSON(http.StatusOK, gin.H{"test": variablesStringBuilder.String()})
	})

	router.Run()
}
