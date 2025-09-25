package extract

import (
	"go/parser"
	"testing"
)

func TestExtractStringLiteral(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantStr   string
		wantFound bool
	}{
		{
			name:      "String extracted",
			code:      `"Extracted string"`,
			wantStr:   `Extracted string`,
			wantFound: true,
		},
		{
			name:      "Even addition is merged",
			code:      `"Extracted " + "string"`,
			wantStr:   `Extracted string`,
			wantFound: true,
		},
		{
			name:      "Odd addition is merged",
			code:      `"Extracted " + "string" + " is combined"`,
			wantStr:   `Extracted string is combined`,
			wantFound: true,
		},
		{
			name:      "Backquotes are removed",
			code:      "`Extracted string`",
			wantStr:   "Extracted string",
			wantFound: true,
		},
		{
			name:      "Multiline text with backquotes are formatted correctly",
			code:      "`This is an multiline\nstring`",
			wantStr:   "This is an multiline\nstring",
			wantFound: true,
		},
		{
			name:      "Backqoutes with qoutes",
			code:      "`This is an \"Test\" abc`",
			wantStr:   "This is an \"Test\" abc",
			wantFound: true,
		},
		{
			name:      "Backqoutes with qoutes",
			code:      `"This is an \"Test\" abc"`,
			wantStr:   "This is an \"Test\" abc",
			wantFound: true,
		},
		{
			name:      "Escaping",
			code:      `"\\caf\u00e9"`,
			wantStr:   `\café`,
			wantFound: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.code)
			if err != nil {
				t.Errorf("Expression %s could not be parsed: %v", tt.code, expr)
			}
			extractedStr, found := StringLiteral(expr)
			if extractedStr != tt.wantStr {
				t.Errorf("StringLiteral() string = %v, want %v", extractedStr, tt.wantStr)
			}
			if (found == nil) == tt.wantFound {
				t.Errorf("StringLiteral() got1 = %v, want %v", found, found == nil)
			}
		})
	}

	t.Run("Nil is ignored", func(t *testing.T) {
		extractedStr, found := StringLiteral(nil)
		if extractedStr != "" {
			t.Errorf("StringLiteral() string = %v, want %v", extractedStr, "")
		}
		if found != nil {
			t.Errorf("StringLiteral() got1 = %v, want %v", found, false)
		}
	})
}
