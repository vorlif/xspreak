package main

import (
	alias "github.com/vorlif/spreak/localize"
)

// TRANSLATORS: Name of the app
//goland:noinspection GoVarAndConstTypeMayBeOmitted
var applicationName alias.Singular = "app"

var ignored = "no localize assign global"

var backtrace = "backtrace"

const (
	// TRANSLATORS: Weekday
	Monday alias.Singular = "monday"
)

const (
	Tuesday          alias.Singular = "tuesday"
	Wednesday        alias.Singular = "wednesday"
	Thursday, Friday alias.Singular = "thursday", "friday"
)
