package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/etype"
)

func TestPrepareKey(t *testing.T) {
	noKey := extract.Issue{IDToken: etype.Singular, MsgID: "msgid", PluralID: "pluralid"}
	key := extract.Issue{IDToken: etype.Key, MsgID: "keyid", PluralID: "keypluralid"}
	pluralKey := extract.Issue{IDToken: etype.PluralKey, MsgID: "id"}

	p := NewPrepareKey()
	res, err := p.Process([]extract.Issue{noKey, key, pluralKey})
	assert.NoError(t, err)
	require.Len(t, res, 3)

	assert.EqualValues(t, noKey, res[0])
	assert.EqualValues(t, key, res[1])

	pluralKey.PluralID = pluralKey.MsgID
	assert.EqualValues(t, pluralKey, res[2])
}
