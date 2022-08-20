package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	// "path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/roboll/helmfile/pkg/maputil"
	"gopkg.in/yaml.v2"
)

var inputVariableFile = flag.String("v", "", "Variables YAML file (optional)")
var sourceFile = flag.String("s", "", "Source file path to parse")
var patternDirectory = flag.String("p", "", "Path to patterns directory (optional)")
var targetFile = flag.String("o", "", "Output file path (optional - writes to standard output otherwise)")

func checkArgs() {
	checkErr := false
	switch *sourceFile {
	case "":
		fmt.Println("Error : No source file provided")
		checkErr = true
		break
	default:
		source, err := os.Open(*sourceFile)
		if err != nil {
			fmt.Println(err)
			checkErr = true
			break
		}
		defer source.Close()
		fileInfo, err := source.Stat()
		if err != nil {
			fmt.Println(err)
			checkErr = true
			break
		}
		if !fileInfo.Mode().IsRegular() {
			fmt.Println("Error : source path provided is not a regular file")
			checkErr = true
			break
		}
	}

	switch *inputVariableFile {
	case "":
		break
	default:
		source, err := os.Open(*inputVariableFile)
		if err != nil {
			fmt.Println(err)
			checkErr = true
			break
		}
		defer source.Close()
		fileInfo, err := source.Stat()
		if err != nil {
			fmt.Println(err)
			checkErr = true
			break
		}
		if !fileInfo.Mode().IsRegular() {
			fmt.Println("Error : variables path provided is not a regular file")
			checkErr = true
			break
		}
	}

	switch *patternDirectory {
	case "":
		break
	default:
		source, err := os.Open(*patternDirectory)
		if err != nil {
			fmt.Println(err)
			checkErr = true
			break
		}
		defer source.Close()
		fileInfo, err := source.Stat()
		if err != nil {
			fmt.Println(err)
			checkErr = true
			break
		}
		if !fileInfo.IsDir() {
			fmt.Println("Error : pattern path provided is not a directory")
			checkErr = true
			break
		}
	}
	if checkErr {
		flag.Usage()
		os.Exit(1)
	}
}

// func indentInclude(fileName string) {
// 	// file, err := os.Open(*inputVariableFile)
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// 	return
// 	// }
// 	// defer file.Close()
// 	content, err := ioutil.ReadFile(fileName)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Convert []byte to string and print to screen
// 	text := string(content)
// 	fmt.Println(text)
// }

func main() {
	flag.Parse()
	checkArgs()

	// indentInclude(*sourceFile)

	var parsedFiles []string
	// parsedFiles = append(parsedFiles, *sourceFile)
	if *patternDirectory != "" {
		patternFolder, err := os.Open(*patternDirectory)
		defer patternFolder.Close()
		if err != nil {
			log.Fatal(err)
		}
		patterns, err := patternFolder.Readdir(-1)
		if err != nil {
			log.Fatal(err)
		}
		for _, pattern := range patterns {
			parsedFiles = append(parsedFiles, *patternDirectory+"/"+pattern.Name())
		}
	}

	//Load variables
	variables := map[string]interface{}{}
	if *inputVariableFile != "" {
		file, err := os.Open(*inputVariableFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fileinfo, err := file.Stat()
		if err != nil {
			fmt.Println(err)
			return
		}
		filesize := fileinfo.Size()
		buffer := make([]byte, filesize)
		file.Read(buffer)

		yaml.Unmarshal(buffer, &variables)
		//Convert decoded yaml value so nested map are all map[string]{interface} instead of map[interface{}]interface{}
		variables, err = maputil.CastKeysToStrings(variables)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	//Set output file
	var output = os.Stdout
	if *targetFile != "" {
		fmt.Printf("Writing to %s ...\n", *targetFile)
		newFile, err := os.Create(*targetFile)
		if err != nil {
			panic(err)
		}
		output = newFile
	}

	//Define the template
	var tpl template.Template

	//Define a custom function using the template
	var funcMap template.FuncMap = map[string]interface{}{}
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := tpl.ExecuteTemplate(buf, name, data); err != nil {
			fmt.Println(err.Error())
			return "", err
		}
		return buf.String(), nil
	}

	//Add custom and sprig lib functions to the template
	tpl.Funcs(sprig.TxtFuncMap()).Funcs(funcMap)

	//Get sourcefile name
	templateName := filepath.Base(*sourceFile)
	//Create the main template from the sourceFile
	b, err := ioutil.ReadFile(*sourceFile)
	if err != nil {
		panic(err)
	}
	tpl.New(templateName).Parse(string(b))

	//Add patterns to template
	_, err = tpl.ParseFiles(parsedFiles...)
	if err != nil {
		panic(err)
	}

	//Applies the main template
	result := bytes.NewBuffer(nil)
	err = tpl.ExecuteTemplate(result, templateName, variables)
	if err != nil {
		panic(err)
	}
	//Check result content
	scanner := bufio.NewScanner(bytes.NewBuffer(result.Bytes()))
	lineIndex := 1
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<no value>") {
			fmt.Fprintf(os.Stderr, "\033[1;33m%s%d%s\033[0m", "Warning: <no value> detected at line ", lineIndex, ",  you may try to access a non existant YAML variable\n")
		}
		lineIndex++
	}

	//Send result to output file
	output.Write(result.Bytes())
}
