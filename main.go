package main

import (
	"github.com/alec-rabold/zipspy/cmd"
	"github.com/spf13/cobra"
)

var (
	// Version is the git version of the code.
	// It is set during build time through the Makefile.
	Version = "unknown"
)

func main() {
	c := cmd.Root()
	c.Version = Version
	cobra.CheckErr(c.Execute())
}
