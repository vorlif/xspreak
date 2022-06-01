package encoder

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/result"
)

func TestJSONEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewJSONEncoder(&buf, "")

	t.Run("key contains context", func(t *testing.T) {
		buf.Reset()

		iss := result.Issue{Context: "ctx", MsgID: "id"}
		err := enc.Encode([]result.Issue{iss})
		require.NoError(t, err)

		want := `{"id_ctx":{"context":"ctx","other":"id"}}
`
		assert.Equal(t, want, buf.String())
	})

	t.Run("by key type text is removed", func(t *testing.T) {
		buf.Reset()

		err := enc.Encode([]result.Issue{
			{MsgID: "s.id", IDToken: etype.Key},
			{MsgID: "p.id", PluralID: "p.id", IDToken: etype.PluralKey},
			{Context: "ctx", MsgID: "sc.id", IDToken: etype.Key},
			{Context: "ctx", MsgID: "pc.id", PluralID: "pc.id", IDToken: etype.PluralKey},
		})
		require.NoError(t, err)
		want := `{"p.id":{"one":"","other":""},"pc.id_ctx":{"context":"ctx","one":"","other":""},"s.id":"","sc.id_ctx":{"context":"ctx","other":""}}
`
		assert.JSONEq(t, want, buf.String())
	})
}

func TestJSONMessage_MarshalJSON(t *testing.T) {
	t.Run("empty returns empty string", func(t *testing.T) {
		msg := make(JSONMessage)
		data, err := json.Marshal(msg)
		require.NoError(t, err)
		assert.Equal(t, `""`, string(data))
	})

	t.Run("single entry returns string", func(t *testing.T) {
		msg := make(JSONMessage)
		msg["other"] = "a"
		data, err := json.Marshal(msg)
		require.NoError(t, err)
		assert.Equal(t, `"a"`, string(data))
	})

	t.Run("keeps order", func(t *testing.T) {
		msg := make(JSONMessage)
		msg["other"] = "a"
		msg["many"] = "b"
		msg["one"] = "c"
		msg["context"] = "d"

		data, err := json.Marshal(msg)
		require.NoError(t, err)
		want := `{"context":"d","one":"c","many":"b","other":"a"}`
		assert.Equal(t, want, string(data))
	})
}

func TestJSONMessage_UnmarshalJSON(t *testing.T) {
	t.Run("string as other", func(t *testing.T) {
		data := []byte(`"a"`)
		var msg JSONMessage

		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Len(t, msg, 1)
		if assert.Contains(t, msg, "other") {
			assert.Equal(t, "a", msg["other"])
		}
	})

	t.Run("object as map", func(t *testing.T) {
		data := []byte(`{"context":"ctx", "one":"b", "other":"a"}`)
		var msg JSONMessage

		require.NoError(t, json.Unmarshal(data, &msg))
		assert.Len(t, msg, 3)
		if assert.Contains(t, msg, "context") {
			assert.Equal(t, "ctx", msg["context"])
		}
		if assert.Contains(t, msg, "one") {
			assert.Equal(t, "b", msg["one"])
		}
		if assert.Contains(t, msg, "other") {
			assert.Equal(t, "a", msg["other"])
		}
	})

	t.Run("empty object", func(t *testing.T) {
		data := []byte(`{}`)
		var msg JSONMessage

		require.NoError(t, json.Unmarshal(data, &msg))
		if assert.Contains(t, msg, "other") {
			assert.Equal(t, "", msg["other"])
		}
	})

	t.Run("empty string", func(t *testing.T) {
		data := []byte(`""`)
		var msg JSONMessage

		require.NoError(t, json.Unmarshal(data, &msg))
		if assert.Contains(t, msg, "other") {
			assert.Equal(t, "", msg["other"])
		}
	})
}
