package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alec-rabold/zipspy/pkg/reader"
	"github.com/alec-rabold/zipspy/pkg/zipspy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func Extract() *cobra.Command {
	var all bool
	var inFiles, outFiles []string
	cmd := &cobra.Command{
		Use:   "extract -file f1.txt [--out out.txt] [--separator \"\\n---\\n\"]",
		Short: "Extract one or more files from the zip archive.",
		Long: `Downloads the specified files from the base zip archive.

Use the "--file/-f" flag to extract one or more files:

	zipspy extract --location file://archive.zip --file myfile.txt 
	zipspy extract --location file://archive.zip -f file1.txt -f file2.txt -file3.txt 

By default, the contents will be written to stdout.
If you wish to separate multiple files' contents, use the "--separator" flag:

	zipspy extract --location file://archive.zip -f file1.txt -f file2.txt --separator "\n---\n" 

You may optionally include the "--out/-o" flag to write the contents to a file instead:

	zipspy extract --location file://archive.zip -f myfile.txt --out destination.txt
	zipspy extract --location file://archive.zip -f file1.txt -f file2.txt -o destination.txt --separator "\n---\n"

If you specify more than one output, each file will be writen to the corresponding desination:

	zipspy extract --location file://archive.zip -f file1.txt -o dest1.txt -f file2.txt -o dest2.txt -f file3.txt -o dest3.txt
`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Parent().PersistentPreRunE(cmd.Parent(), args); err != nil {
				return fmt.Errorf("root pre-run failed: %v", err)
			}
			if err := validateExtractCommand(cmd); err != nil {
				cmd.Usage()
				return fmt.Errorf("validation failed: %v", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			zip, err := zipspy.NewClient(cfg.zipReader)
			if err != nil {
				return fmt.Errorf("failed to create zipspy client: %v", err)
			}

			files := getFiles(cmd, zip)
			if len(inFiles) != 0 && len(inFiles) != len(files) {
				if len(outFiles) > 1 {
					return fmt.Errorf("number of input files must match number of found files in order to write to multiple files (input: %d) (found: %d)", len(inFiles), len(files))
				}
				log.Warnf("number of input files does not match number of found files (input: %d) (found: %d)", len(inFiles), len(files))
			}

			outFile := os.Stdout
			if len(outFiles) == 1 {
				outFile, err = os.OpenFile(outFiles[0], os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return fmt.Errorf("failed to open file (name: %s): %w", outFiles[0], err)
				}
			}
			defer outFile.Close()
			// TODO: parallelize with variable number of workers here
			for idx, file := range files {
				// Kind of hacky, but skip if directory
				if strings.HasSuffix(file.Name, "/") {
					continue
				}
				rc, err := file.Open()
				if err != nil {
					return fmt.Errorf("failed to open file (name: %s): %w", file.Name, err)
				}
				defer rc.Close()

				if len(outFiles) > 1 {
					outFile, err = os.OpenFile(outFiles[idx], os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						return fmt.Errorf("failed to open file (name: %s): %w", outFiles[0], err)
					}
					defer outFile.Close()
				}
				if err := writeToFile(bufio.NewReader(rc), bufio.NewWriter(outFile), buildSeparator(cmd)); err != nil {
					return fmt.Errorf("failed writing contents to file: %w", err)
				}
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringSliceVarP(&inFiles, "file", "f", []string{}, "(required) names of the files/paths to extract (e.g. plan.txt, /path/to/plan.txt, /directory)")
	cmd.PersistentFlags().StringSliceVarP(&outFiles, "out", "o", []string{}, "(optional) name(s) of the file(s) to write output to")
	cmd.PersistentFlags().String("separator", "", "(optional) separator when combining the output of multiple files")
	cmd.PersistentFlags().BoolVar(&all, "all", false, "(optional) whether to extract all files in the zip archive")
	cmd.PersistentFlags().Bool("no-newlines", false, "(optional) omit the newlines appended to files")
	return cmd
}

func validateExtractCommand(cmd *cobra.Command) error {
	files, err := cmd.Flags().GetStringSlice("file")
	if err != nil {
		return fmt.Errorf("at least one file must be specified, or use the --all flag: %v", err)
	}
	outfiles, _ := cmd.Flags().GetStringSlice("out")
	if len(outfiles) > 1 && (len(outfiles) != len(files)) {
		return fmt.Errorf("one output file must be specified for each search term, or you may use a single output file")
	}
	return nil
}

func getFiles(cmd *cobra.Command, zip *zipspy.Client) []*reader.File {
	all, _ := cmd.Flags().GetBool("all")
	inFiles, _ := cmd.Flags().GetStringSlice("file")
	if all {
		return zip.AllFiles()
	}
	return zip.GetFiles(inFiles)
}

func writeToFile(r *bufio.Reader, w *bufio.Writer, separator string) error {
	buf := make([]byte, 128)
	for {
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := w.Write(buf[:n]); err != nil {
			return err
		}
	}
	if _, err := w.WriteString(separator); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func buildSeparator(cmd *cobra.Command) string {
	separator, _ := cmd.Flags().GetString("separator")
	noNewlines, _ := cmd.Flags().GetBool("no-newlines")
	if noNewlines {
		return separator
	}
	if separator == "" {
		return "\n"
	}
	return "\n" + separator + "\n"
}
