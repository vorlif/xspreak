package processors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReGoStringFormat(t *testing.T) {
	tests := []struct {
		text   string
		assert assert.BoolAssertionFunc
	}{
		{"", assert.False},
		{"text", assert.False},
		{"text %%", assert.False},
		{"1234 %%", assert.False},
		{"%v", assert.True},
		{"%#v", assert.True},
		{"%T", assert.True},
		{"%+v", assert.True},
		{"%t", assert.True},
		{"%b", assert.True},
		{"%c", assert.True},
		{"%d", assert.True},
		{"%o", assert.True},
		{"%O", assert.True},
		{"%x", assert.True},
		{"%X", assert.True},
		{"%U", assert.True},
		{"%e", assert.True},
		{"%E", assert.True},
		{"%g", assert.True},
		{"%G", assert.True},
		{"%s", assert.True},
		{"%q", assert.True},
		{"%p", assert.True},
		{"%f", assert.True},
		{"%9f", assert.True},
		{"%.2f", assert.True},
		{"%9.2f", assert.True},
		{"%9.f", assert.True},
		{"%[2]d", assert.True},
		// {"1234 %%s", assert.False},
		{"%#[1]x", assert.True},
		{"%*[2]d", assert.True},
		{"%.[2]d", assert.True},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			tt.assert(t, reGoStringFormat.MatchString(tt.text))
		})
	}
}
