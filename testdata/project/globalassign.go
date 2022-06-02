package main

import (
	alias "github.com/vorlif/spreak/localize"
)

//goland:noinspection GoVarAndConstTypeMayBeOmitted
// TRANSLATORS: Name of the app
var applicationName = "app"

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
