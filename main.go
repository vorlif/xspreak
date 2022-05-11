package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"github.com/vorlif/xspreak/commands"
)

func init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{DisableTimestamp: true})
	logrus.SetLevel(logrus.InfoLevel)
}

func main() {
	executor := commands.NewExecutor()
	if err := executor.Execute(); err != nil {
		logrus.Warn(err)
	}
}
