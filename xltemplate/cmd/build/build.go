package build

import (
	"do3b/xltemplate/api/templateengine"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/roboll/helmfile/pkg/maputil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type buildFlags struct {
	variables string
	source    string
	patterns  string
	output    string
}

// NewCmdVersion makes a new version command.
func NewCmdVersion(w io.Writer) *cobra.Command {
	opts := &buildFlags{}

	cmd := cobra.Command{
		Use:     "build",
		Short:   "Build a template file",
		Long:    `Build a template file from a source file, patterns and a set of variables.`,
		Example: `xltemplate build`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(*opts, w)
		},
	}

	cmd.Flags().StringVar(&opts.variables, "variables", "", "Variables YAML file")
	cmd.Flags().StringVar(&opts.source, "source", "", "Source file path to parse")
	cmd.Flags().StringVar(&opts.patterns, "patterns", "", "Path to patterns directory")
	cmd.Flags().StringVar(&opts.output, "output", "", "Output file path (optional - writes to standard output otherwise)")
	return &cmd
}

func Run(opts buildFlags, w io.Writer) error {
	variables := map[string]interface{}{}
	if opts.variables != "" {
		file, err := os.Open(opts.variables)
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

		yaml.Unmarshal(buffer, &variables)
		//Convert decoded yaml value so nested map are all map[string]{interface} instead of map[interface{}]interface{}
		variables, err = maputil.CastKeysToStrings(variables)
		if err != nil {
			fmt.Println(err)
		}
	}

	source := ""
	if opts.source != "" {
		b, err := os.ReadFile(opts.source)
		if err != nil {
			panic(err)
		}
		source = string(b)
	}

	result := templateengine.Parse(opts.source, variables, source, readPatternDirectory(opts.patterns))
	var output = os.Stdout
	if opts.output != "" {
		slog.Info("Writing to file", "file", opts.output)
		newFile, err := os.Create(opts.output)
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
