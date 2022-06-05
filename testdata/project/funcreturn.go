package main

import (
	"math/rand"

	"github.com/vorlif/spreak/localize"
	alias "github.com/vorlif/spreak/localize"
)

const msgB = "bt_msgB"

func returnSingle() localize.MsgID {
	return "single"
}

func returnPlural() (alias.MsgID, localize.Plural) {
	return "plural_s", "plural_p"
}

func returnContext() (alias.MsgID, localize.Context, localize.Plural) {
	return "context_s", "context_c", "context_p"
}

func returnFull() (alias.MsgID, alias.Context, alias.Plural, alias.Domain) {
	return "full_s", "full_c", "full_p", "full_d"
}

func returnBacktracking() (alias.MsgID, alias.Context, alias.Plural, alias.Domain) {
	ctxA := "bt_ctx_a"
	ctxB := "bt_ctx_b"

	domainA := "bt_domain_A"
	domainB := "bt_domain_B"

	msgA := "bt_msgA"

	pluralA := "bt_plural_a"
	pluralB := "bt_plural_b"

	switch rand.Intn(101) {
	case 1:
		return msgA, "bt_ctx_c", pluralA, "bt_domain_c"
	case 2:
		return msgB, ctxA, pluralB, domainA
	default:
		return "bt_msg_c", ctxB, pluralA, domainB
	}
}
