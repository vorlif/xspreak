package main

import (
	"fmt"

	"github.com/vorlif/spreak/localize"
)

const (
	constCtx = "constCtxVal"
)

type M struct {
	Test  localize.Singular
	Hello string
}

func main() {

	// test comment
	fmt.Println(localize.Singular("init"))
}
