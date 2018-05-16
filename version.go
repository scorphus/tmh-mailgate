package main

import (
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version info",
		Run: func(cmd *cobra.Command, args []string) {
			version()
		},
	}

	Version  string
	Revision string
	DateTime string
	Message  string

	StartTime      time.Time
	ConfigLoadTime time.Time
)

func init() {
	if Version == "" {
		Version = "unknown"
	}
	if Revision == "" {
		Revision = "unknown"
	}
	if DateTime == "" {
		DateTime = "unknown"
	}
	if Message == "" {
		Message = "unknown"
	}
	StartTime = time.Now()
	rootCmd.AddCommand(versionCmd)
}

func version() {
	mainlog.Infof("tmh-mailgate %s built with %s", Version, runtime.Version())
	mainlog.Infof("Revision: %s", Revision)
	mainlog.Infof("DateTime: %s", DateTime)
	mainlog.Infof("Message:  %s", Message)
}
