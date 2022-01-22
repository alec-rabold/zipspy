package cmd

import (
	"bufio"
	"io"

	"github.com/spf13/cobra"
)

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
