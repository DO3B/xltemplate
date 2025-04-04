package build

import (
	"do3b/xltemplate/api/loader"
	"do3b/xltemplate/api/templateengine"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/imdario/mergo"
	"github.com/roboll/helmfile/pkg/maputil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type buildFlags struct {
	Variables string
	Source    string
	Patterns  []string
	Output    string
}

// NewCmdVersion makes a new version command.
func NewCmdVersion(fileSystem filesys.FileSystem, w io.Writer) *cobra.Command {
	opts := buildFlags{}

	cmd := cobra.Command{
		Use:     "build",
		Short:   "Build a template file",
		Long:    `Build a template file from a source file, patterns and a set of variables.`,
		Example: `xltemplate build xltemplate.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			xltemplateFile := buildFlags{}
			if len(args) > 0 && args[0] != "" {
				file, err := os.Open(args[0])
				if err != nil {
					fmt.Println(err)
				}
				defer file.Close()
				fileinfo, err := file.Stat()
				if err != nil {
					fmt.Println(err)
				}
				filesize := fileinfo.Size()
				buffer := make([]byte, filesize)
				file.Read(buffer)

				yaml.Unmarshal(buffer, &xltemplateFile)
				slog.Debug("Xltemplate file content", "xltemplateFile", xltemplateFile)
			}

			// Merging the content of the xltemplate file with the command line arguments
			if err := mergo.Merge(&opts, xltemplateFile, mergo.WithAppendSlice); err != nil {
				slog.Error("Error merging xltemplate file with command line arguments", "error", err)
			}

			slog.Debug("Executing build command with options", "opts", opts)
			return Run(opts, fileSystem, w)
		},
	}

	cmd.Flags().StringVar(&opts.Variables, "variables", "", "variables YAML file")
	cmd.Flags().StringVar(&opts.Source, "source", "", "source file path to parse")
	cmd.Flags().StringArrayVar(&opts.Patterns, "patterns", []string{}, "path to patterns directory")
	cmd.Flags().StringVar(&opts.Output, "output", "", "output file path (optional - writes to standard output otherwise)")
	return &cmd
}

func Run(opts buildFlags, fileSystem filesys.FileSystem, w io.Writer) error {
	var err error

	patterns := []string{}
	pattern_loaders := []loader.FileLoader{}
	for _, pattern := range opts.Patterns {
		pattern_loader, err := loader.NewLoader(
			loader.RestrictionNone,
			pattern,
			fileSystem,
		)
		if err != nil {
			slog.Error("Error loading patterns", "error", err)
			return err
		}

		patterns = append(patterns, readPatternDirectory(pattern_loader.Root())...)
		pattern_loaders = append(pattern_loaders, *pattern_loader)
	}

	variables := map[string]interface{}{}
	if opts.Variables != "" {
		variables, err = loadYamlFromFile(opts.Variables)
		if err != nil {
			slog.Error("Error loading variables", "error", err)
		}

		if variable, exists := variables[":includes"]; exists {
			// Convert decoded yaml value so nested map are all map[string]{interface} instead of map[interface{}]interface{}
			slog.Debug("Includes found in variables", ":includes", variable)
			var includedVariables []map[string]interface{}
			if list, ok := variable.([]interface{}); ok {
				for _, item := range list {
					includedVariable, err := loadYamlFromFile(item.(string))
					includedVariables = append(includedVariables, includedVariable)
					if err != nil {
						slog.Error("Error loading included variables", "error", err)
					}
				}
			} else {
				slog.Error("Includes must be a list", "includes", variable)
				fmt.Println("Includes must be a list")
			}

			for _, includedVariable := range includedVariables {
				if err := mergo.Merge(&variables, includedVariable); err != nil {
					slog.Error("Error merging included variables", "error", err)
				}
			}

			delete(variables, ":includes")
			slog.Debug("Merged variables", "variables", variables)
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	source := ""
	if opts.Source != "" {
		source_loader, err := loader.NewLoader(
			loader.RestrictionNone,
			opts.Source,
			fileSystem,
		)
		if err != nil {
			panic(err)
		}
		b, err := source_loader.Load(source_loader.FilePath)
		if err != nil {
			panic(err)
		}
		source = string(b)
	}

	templateEngine := templateengine.NewTemplateEngine(opts.Source, variables, source, patterns)
	result, err := templateEngine.Parse()
	for _, pattern_loader := range pattern_loaders {
		pattern_loader.Cleanup()
	}
	if err != nil {
		panic(err)
	}

	var output = os.Stdout
	if opts.Output != "" {
		slog.Info("Writing to file", "file", opts.Output)
		newFile, err := os.Create(opts.Output)
		if err != nil {
			panic(err)
		}
		output = newFile
	}

	output.Write([]byte(result))

	return nil
}

func recursivelyReadPatternDirectory(path string, dirEntry fs.DirEntry, patterns []string) []string {
	fileInfo, err := dirEntry.Info()
	if err != nil {
		slog.Error("Error getting file info", "error", err)
	}

	if fileInfo.IsDir() {
		return append(patterns, readPatternDirectory(path+"/"+fileInfo.Name())...)
	} else {
		return append(patterns, path+"/"+fileInfo.Name())
	}
}

func readPatternDirectory(path string) []string {
	patternFolder, err := os.Open(path)
	if err != nil {
		slog.Error("Error opening pattern directory", "error", err)
	}
	defer patternFolder.Close()

	patterns, err := patternFolder.ReadDir(-1)
	if err != nil {
		slog.Error("Error reading pattern directory", "error", err)
	}

	var parsedFiles []string
	for _, pattern := range patterns {
		parsedFiles = recursivelyReadPatternDirectory(path, pattern, parsedFiles)
	}

	return parsedFiles
}

func loadYamlFromFile(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	result, err = maputil.CastKeysToStrings(result)
	if err != nil {
		return nil, fmt.Errorf("failed to cast keys to strings: %w", err)
	}
	return result, nil
}
