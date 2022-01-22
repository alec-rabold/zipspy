package cmd

import (
	"fmt"
	"os"

	"github.com/alec-rabold/zipspy/pkg/provider"
	"github.com/alec-rabold/zipspy/pkg/provider/aws/s3"
	"github.com/alec-rabold/zipspy/pkg/provider/local"
	"github.com/alec-rabold/zipspy/pkg/zipspy"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cfg config

type config struct {
	development     bool
	archiveLocation string
	zipReader       zipspy.Reader
}

// Root returns the cobra.Command containing all child commands and sets global flags.
func Root() *cobra.Command {
	var verbosity string
	cmd := &cobra.Command{
		Use:   "zipspy",
		Short: "Interface with remote ZIP archives",
		Long: `                       
 ____  __  ____  ____  ____  _  _ 
(__  )(  )(  _ \/ ___)(  _ \( \/ )
 / _/  )(  ) __/\___ \ ) __/ )  / 
(____)(__)(__)  (____/(__)  (__/    

Zipspy allows you interact with ZIP archives stored in remote locations without
requiring a local copy. For example, you can list the filenames in an S3 ZIP archive, 
download a subset of files, search and retrieve files with regular expressions, and more!`,
	}
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := cfg.initProvider(); err != nil {
			return fmt.Errorf("failed to initialize provider: %v", err)
		}
		if err := setupLogger(verbosity); err != nil {
			return fmt.Errorf("failed to initialize logger: %v", err)
		}
		return nil
	}
	cmd.PersistentFlags().BoolVar(&cfg.development, "development", false, "Whether or not to use development settings")
	cmd.PersistentFlags().StringVar(&cfg.archiveLocation, "location", "", `Protocol and address of your ZIP archive ("file://archive.zip", "s3://<bucket_name>/archive.zip")`)
	cmd.LocalFlags().StringVar(&verbosity, "verbosity", logrus.WarnLevel.String(), "Global log level (trace, debug, info, warn, error, fatal, panic")
	must(cmd.MarkPersistentFlagRequired("location"))

	cmd.AddCommand(Extract())

	return cmd
}

func (c *config) initProvider() error {
	if c.archiveLocation == "" {
		return fmt.Errorf("location must not be empty")
	}
	r := provider.NewRegistry(
		provider.WithProvider("s3", "s3://", s3.NewClient),
		provider.WithProvider("local", "file://", local.NewClient),
	)
	zr, err := r.GetPlugin(c.archiveLocation)
	if err != nil {
		return fmt.Errorf("failed to get plugin for location %s: %w", c.archiveLocation, err)
	}
	c.zipReader = zr
	return nil
}

func setupLogger(verbosity string) error {
	log.SetOutput(os.Stdout)
	level, err := logrus.ParseLevel(verbosity)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	if cfg.development && level < log.DebugLevel {
		log.SetLevel(log.DebugLevel)
	}
	return nil
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
