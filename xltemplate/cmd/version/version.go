package version

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewCmdVersion makes a new version command.
func NewCmdVersion(w io.Writer) *cobra.Command {
	versionCmd := cobra.Command{
		Use:     "version",
		Short:   "Prints the xltemplate version",
		Example: `xltemplate version`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(w)
		},
	}
	return &versionCmd
}

func Run(w io.Writer) error {
	fmt.Fprintln(w, "v2.0.0")
	return nil
}
