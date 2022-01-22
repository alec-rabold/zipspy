package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/alec-rabold/zipspy/pkg/zipspy"
	"github.com/spf13/cobra"
)

func List() *cobra.Command {
	var outFileName string
	var includeDirectoryNames bool
	cmd := &cobra.Command{
		Use:   "list [--include-directory-names]",
		Short: "List all file names from a zip archive.",
		Long: `Prints out the names of all files contained within a zip archive.
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return fmt.Errorf("root pre-run failed: %v", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zip, err := zipspy.NewClient(cfg.zipReader)
			if err != nil {
				return fmt.Errorf("failed to create zipspy client: %v", err)
			}

			outFile := os.Stdout
			if outFileName != "" {
				outFile, err = os.OpenFile(outFileName, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to open file (name: %s): %w", outFileName, err)
				}
				defer outFile.Close()
			}
			for _, file := range zip.AllFiles() {
				// Skip directory names from file name list (e.g. "my/dir/")
				if !includeDirectoryNames && strings.HasSuffix(file.Name, "/") {
					continue
				}
				r := strings.NewReader(file.Name)
				if err := writeToFile(bufio.NewReader(r), bufio.NewWriter(outFile), buildSeparator(cmd)); err != nil {
					return fmt.Errorf("failed writing contents to file: %w", err)
				}
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&outFileName, "out", "o", "", "(optional) name of a file to write output to")
	cmd.PersistentFlags().BoolVar(&includeDirectoryNames, "include-directory-names", false, "(optional) include the leaf names of directories")
	cmd.PersistentFlags().String("separator", "", "(optional) separator when combining the output of multiple file names")
	cmd.PersistentFlags().Bool("no-newlines", false, "(optional) omit the newlines appended to file names")
	return cmd
}
