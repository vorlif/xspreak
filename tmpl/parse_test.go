package tmpl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	text := `This is an test {{.T.Get "hello" "world"}}
{{.T.Get "foo"}}

`
	res, err := ParseBytes("test", []byte(text))
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.OffsetLookup, 4)
	assert.Equal(t, 1, res.OffsetLookup[0].Line)
}

func TestExtractComments(t *testing.T) {
	text := `{{/* start comment */}} This is an test {{.T.Get "hello" "world"}}

{{/*a comment 
with
multiline */}}
{{.T.Get "foo"}}

{{- /*   also a comment */ -}}
`
	res, err := ParseBytes("test", []byte(text))
	assert.NoError(t, err)
	require.NotNil(t, res)

	assert.Len(t, res.Comments, 0)
	res.ExtractComments()
	assert.Len(t, res.Comments, 3)
}

func TestParseHtml(t *testing.T) {
	text, err := os.ReadFile("../testdata/tmpl/five.tmpl")
	if err != nil {
		panic(err)
	}

	res, err := ParseBytes("test", text)
	assert.NoError(t, err)
	require.NotNil(t, res)
}
