package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/extract"
)

func TestSkipEmptyMsgID(t *testing.T) {
	issuses := []extract.Issue{
		{PluralID: "p", Context: "ctx"},
		{MsgID: "id"},
	}

	p := NewSkipEmptyMsgID()
	res, err := p.Process(issuses)
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.EqualValues(t, extract.Issue{MsgID: "id"}, res[0])
}
