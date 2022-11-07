package merger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/spreak/catalog/cldrplural"
)

func TestMergeJson(t *testing.T) {

	t.Run("missing IDs will be created", func(t *testing.T) {
		src := []byte(`{
"a": "", "c_ctx": {"context": "ctx", "other": "c"}, "d": "",  "b": ""
}`)
		dst := []byte(`{
"a": "A",
"d": "D"
}`)
		res := MergeJSON(src, dst, []cldrplural.Category{cldrplural.One, cldrplural.Many, cldrplural.Other})
		require.NotNil(t, res)

		want := `{
  "a": "A",
  "b": "",
  "c_ctx": {
    "context": "ctx",
    "other": "c"
  },
  "d": "D"
}`
		assert.JSONEq(t, want, string(res))
	})

	t.Run("new plurals are created", func(t *testing.T) {
		src := []byte(`{
"a": {"zero": "", "other": "O"},
"b_ctx": {"context": "ctx", "zero": "", "other": ""}
}`)

		res := MergeJSON(src, nil, []cldrplural.Category{cldrplural.One, cldrplural.Many, cldrplural.Other})
		require.NotNil(t, res)

		want := `{
"a": {"one": "", "many": "", "other": "O"},
"b_ctx": {"context": "ctx", "one": "", "many": "", "other": ""}
}`
		assert.JSONEq(t, want, string(res))
	})
}
