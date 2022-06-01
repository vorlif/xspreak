package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vorlif/xspreak/result"
)

func TestSkipEmptyMsgID(t *testing.T) {
	issuses := []result.Issue{
		{PluralID: "p", Context: "ctx"},
		{MsgID: "id"},
	}

	p := NewSkipEmptyMsgID()
	res, err := p.Process(issuses)
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.EqualValues(t, result.Issue{MsgID: "id"}, res[0])
}
