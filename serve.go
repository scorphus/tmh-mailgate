package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/flashmob/go-guerrilla"
	"github.com/flashmob/go-guerrilla/log"
	"github.com/spf13/cobra"
)

const (
	defaultPidFile = "/var/run/tmh-mailgate.pid"
)

var (
	configPath string
	pidFile    string
	serveCmd   = &cobra.Command{
		Use:   "serve",
		Short: "Start tmh-mailgate server",
		Run:   serve,
	}
	signalChannel = make(chan os.Signal, 1)
	mainlog       log.Logger
	d             guerrilla.Daemon
)

func init() {
	var err error
	mainlog, err = log.GetLogger(log.OutputStderr.String(), log.InfoLevel.String())
	if err != nil {
		mainlog.WithError(err).Errorf("Failed creating a logger to %s", log.OutputStderr)
	}
	serveCmd.PersistentFlags().StringVarP(
		&configPath,
		"config",
		"c",
		"tmh-mailgate.conf",
		"Path to the configuration file",
	)
	serveCmd.PersistentFlags().StringVarP(
		&pidFile,
		"pidFile",
		"p",
		"tmh-mailgate.pid",
		"Path to the pid file",
	)
	rootCmd.AddCommand(serveCmd)
}

func sigHandler() {
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGUSR1,
	)
	for sig := range signalChannel {
		if sig == syscall.SIGHUP {
			d.ReloadConfigFile(configPath)
			continue
		} else if sig == syscall.SIGUSR1 {
			d.ReopenLogs()
			continue
		} else if sig == syscall.SIGTERM || sig == syscall.SIGQUIT || sig == syscall.SIGINT {
			mainlog.Infof("Shutdown signal caught")
			d.Shutdown()
			mainlog.Infof("Shutdown completed, exiting.")
			break
		}
		mainlog.Infof("Shutdown, unknown signal caught")
		break
	}
}

func serve(cmd *cobra.Command, args []string) {
	version()
	d = guerrilla.Daemon{Logger: mainlog}
	d.AddProcessor("OpsGenie", OpsGenieProcessor)
	err := readConfig(configPath, pidFile)
	if err != nil {
		mainlog.WithError(err).Fatal("Error while reading config")
	}
	fileLimit := getFileLimit()
	if fileLimit > 0 {
		maxClients := 0
		for _, s := range d.Config.Servers {
			maxClients += s.MaxClients
		}
		if maxClients > fileLimit {
			mainlog.Fatalf(
				"Combined max clients (%d) greater than open file limit (%d). "+
					"Increase open file limit or decrease max clients.",
				maxClients,
				fileLimit,
			)
		}
	}
	err = d.Start()
	if err != nil {
		mainlog.WithError(err).Error("Error(s) when starting server(s)")
		os.Exit(1)
	}
	sigHandler()
}

// CmdConfig is a superset of AppConfig with additional options for the CLI
type CmdConfig struct {
	guerrilla.AppConfig
}

func (c *CmdConfig) emitChangeEvents(oldConfig *CmdConfig, app guerrilla.Guerrilla) {
	c.AppConfig.EmitChangeEvents(&oldConfig.AppConfig, app)
}

func readConfig(path string, pidFile string) error {
	if _, err := d.LoadConfig(path); err != nil {
		return err
	}
	if len(pidFile) > 0 {
		d.Config.PidFile = pidFile
	} else if len(d.Config.PidFile) == 0 {
		d.Config.PidFile = defaultPidFile
	}
	if len(d.Config.AllowedHosts) == 0 {
		return errors.New("Empty `allowed_hosts` is not allowed")
	}
	return nil
}

func getFileLimit() int {
	cmd := exec.Command("ulimit", "-n")
	out, err := cmd.Output()
	if err != nil {
		return -1
	}
	limit, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return -1
	}
	return limit
}
