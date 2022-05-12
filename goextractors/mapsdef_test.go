package goextractors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapsDefExtractor(t *testing.T) {
	issues := runExtraction(t, testdataDir, NewMapsDefExtractor())
	assert.NotEmpty(t, issues)

	got := collectIssueStrings(issues)
	want := []string{
		"globalKeyMap-a", "globalKeyMap-b",
		"globalValueMap-a", "globalValueMap-b",
		"globalMap-ka", "globalMap-va", "globalMap-kb", "globalMap-vb",

		"localKeyMap-a", "localKeyMap-b",
		"localValueMap-a", "localValueMap-b",
		"localMap-ka", "localMap-va", "localMap-kb", "localMap-vb",

		"map struct msgid", "map struct plural",
		"map pointer struct msgid", "map pointer struct plural",
	}
	assert.ElementsMatch(t, want, got)
}
