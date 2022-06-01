package tmpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultKeywords(t *testing.T) {
	keywords := DefaultKeywords("..T.", false)
	assert.Len(t, keywords, 32)
	pointCount := 0
	dollarCount := 0
	formatCount := 0
	for _, kw := range keywords {
		if strings.HasPrefix(kw.Name, ".") {
			pointCount++
		}
		if strings.HasPrefix(kw.Name, "$") {
			dollarCount++
		}
		if strings.HasSuffix(kw.Name, "f") {
			formatCount++
		}
	}

	assert.Equal(t, 16, pointCount)
	assert.Equal(t, 16, dollarCount)
	assert.Equal(t, 16, formatCount)
}
