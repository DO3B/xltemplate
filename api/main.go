package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/gin-gonic/gin"
)

type ProspectiveTemplate struct {
	Cia        string      `json:"cia" binding:"required"`
	Components []Component `json:"components" binding:"required"`
}

type Component struct {
	Type         string   `json:"type"`
	Deploy       string   `json:"deploy"`
	Release      string   `json:"release"`
	Environments []string `json:"environments"`
	Packages     []string `json:"packages"`
}

type Pack struct {
	Name  string
	AppId string
}

type Environment struct {
	Name         string
	Trigram      string
	TemplateName string
}

type ComputedComponent struct {
	Type         string
	Deploy       Configuration
	Release      Configuration
	Packages     []Pack
	Environments []Environment
}

type Configuration struct {
	Configuration string
	Account       string
}

func main() {
	router := gin.Default()

	router.POST("/createTemplate", CreateTemplate)

	router.Run()
}

// CreateTemplate godoc
// @Summary Create a template
// @Accept json
// @Produce json
func CreateTemplate(context *gin.Context) {
	var prospectiveTemplate ProspectiveTemplate
	if err := context.BindJSON(&prospectiveTemplate); err != nil {
		context.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	checkPackages, _ := strconv.ParseBool(context.DefaultQuery("checkPackages", "false"))
	// String builder to return a final yaml
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
		templateName := filepath.Base("sample/components/" + patternName + ".tmpl")
		// Create the template based on pattern name
		b, err := ioutil.ReadFile("sample/components/" + patternName + ".tmpl")
		if err != nil {
			templateName = "sample/components/generic.tmpl"
			fmt.Println("No template associated to the component, I will try to create a generic one")
			b, err = ioutil.ReadFile(templateName)
			if err != nil {
				panic(err)
			}
		}
		componentTpl.New(templateName).Parse(string(b))

		var packages []Pack
		for _, pack := range component.Packages {
			if checkPackages {
				fmt.Printf("Checking %s on %s", pack, component.Deploy)
			}
			newPack := Pack{
				Name:  strings.Split(pack, "/")[len(strings.Split(pack, "/"))-1],
				AppId: pack,
			}
			packages = append(packages, newPack)
		}

		var environments []Environment
		for _, env := range component.Environments {
			trigram := cases.Upper(language.Und).String(env[0:3])
			newEnvironment := Environment{
				Name:         env,
				Trigram:      trigram,
				TemplateName: "[" + component.Type + "] " + cases.Title(language.Und).String(env) + " (" + trigram + ")",
			}
			environments = append(environments, newEnvironment)
		}

		deployConfiguration := Configuration{
			Configuration: "XLD-" + component.Deploy,
		}

		releaseConfiguration := Configuration{
			Configuration: "XLR-" + component.Release,
			Account:       "SVCPBIS10BIXLR" + component.Release,
		}

		computedComponent := ComputedComponent{
			Type:         component.Type,
			Deploy:       deployConfiguration,
			Release:      releaseConfiguration,
			Packages:     packages,
			Environments: environments,
		}

		result = bytes.NewBuffer(nil)

		err = componentTpl.ExecuteTemplate(result, templateName, computedComponent)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		variablesStringBuilder.WriteString(result.String() + "\r\n")
	}
	fmt.Println(variablesStringBuilder.String())
	context.JSON(http.StatusOK, variablesStringBuilder.String())
}
