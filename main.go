package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/vorlif/xspreak/commands"
)

// Version can be set at link time.
var Version = "0.0.0"

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	executor := commands.NewExecutor(Version)
	if err := executor.Execute(); err != nil {
		logrus.Warn(err)
	}
}
