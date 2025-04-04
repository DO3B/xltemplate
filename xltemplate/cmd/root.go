/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"do3b/xltemplate/cmd/build"
	"do3b/xltemplate/cmd/version"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var Verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "xltemplate",
	Short: "A quick utility to build template",
	Long: `xltemplate is an utility which allows to build template files
from a source file, patterns and a set of variables.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if Verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	fileSystem := filesys.MakeFsOnDisk()

	rootCmd.AddCommand(
		build.NewCmdVersion(fileSystem, os.Stdout),
		version.NewCmdVersion(os.Stdout),
	)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.xltemplate.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
