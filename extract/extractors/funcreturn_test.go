package extractors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncReturnExtractor(t *testing.T) {
	issues := runExtraction(t, testdataDir, NewFuncReturnExtractor())
	assert.NotEmpty(t, issues)

	want := []string{
		"single",
		"plural_s", "plural_p",
		"context_s", "context_c", "context_p",
		"full_s", "full_c", "full_p", "full_d",

		// backtracking
		"bt_ctx_a", "bt_ctx_b",
		"bt_domain_A", "bt_domain_B",
		"bt_msgA",
		"bt_plural_a", "bt_plural_b",
		"bt_ctx_c", "bt_domain_c",
		"bt_msg_c", "bt_msgB",
		"bt_plural_a",
	}
	got := collectIssueStrings(issues)
	assert.ElementsMatch(t, want, got)
	assert.Len(t, issues, 7)
}
