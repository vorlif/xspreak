package goextractors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncCallExtractor(t *testing.T) {
	issues := runExtraction(t, testdataDir, NewFuncCallExtractor())
	assert.NotEmpty(t, issues)

	want := []string{
		"f-msgid", "f-plural", "f-context", "f-domain",
		"init", "localizer func call",
		"noop-msgid", "noop-plural", "noop-context", "noop-domain",
		"msgid",
		"msgid-n", "pluralid-n",
		"domain-d", "msgid-d",
		"domain-dn", "msgid-dn", "pluralid-dn",
		"context-pg", "msgid-pg",
		"context-np", "msgid-np", "pluralid-np",
		"domain-dp", "context-dp", "singular-dp",
		"domain-dnp", "context-dnp", "msgid-dnp", "pluralid-dnp",
		"submsgid", "subplural", "foo test",
		"generic-call",
		"pre-variadic", "variadic-a", "variadic-b",
		"no-param-msgid", "no-param-plural",
		"multi-names-a", "multi-names-b",
		"init backtrace", "assign backtrace",
		"inline function",

		"constCtxMsg", "constCtxVal",

		"struct-method-call", "generic-struct-method-call",
	}
	got := collectIssueStrings(issues)
	assert.ElementsMatch(t, want, got)

	for _, iss := range issues {
		switch iss.MsgID {
		case "constCtxMsg":
			assert.Equal(t, "constCtxVal", iss.Context)
		}
	}
}
