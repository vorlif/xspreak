package main

import "github.com/sirupsen/logrus"

// This file contains imports of packages that are not interesting for us.

func log() {
	logrus.Info("Logging from uninteresting_imports.go")
}
