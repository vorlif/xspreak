package tmpl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	text := `This is an test {{.T.Get "hello" "world"}}"`
	res, err := ParseBytes("test", []byte(text))
	require.NoError(t, err)
	require.NotNil(t, res)
}
