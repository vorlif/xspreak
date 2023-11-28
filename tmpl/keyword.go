package tmpl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vorlif/xspreak/extract/etype"
)

type Keyword struct {
	Name        string
	IDToken     etype.Token
	SingularPos int
	PluralPos   int
	ContextPos  int
	DomainPos   int
}

func ParseKeywords(spec string, isMonolingual bool) (*Keyword, error) {
	var functionName string
	var args []string
	idx := strings.IndexByte(spec, ':')
	if idx >= 0 {
		functionName = spec[:idx]
		args = strings.Split(spec[idx+1:], ",")
	} else {
		functionName = spec
	}

	k := &Keyword{
		Name:        functionName,
		IDToken:     etype.Singular,
		SingularPos: 0,
		PluralPos:   -1,
		ContextPos:  -1,
		DomainPos:   -1,
	}

	if isMonolingual {
		k.IDToken = etype.Key
	}

	inputType := 0
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}

		lastSign := arg[len(arg)-1]
		if lastSign == 'c' || lastSign == 'd' {
			val, err := strconv.Atoi(arg[:len(arg)-1])
			if err != nil {
				return nil, fmt.Errorf("bad keyword number: %s %w", arg, err)
			}
			if lastSign == 'c' {
				k.ContextPos = val - 1
			} else {
				k.DomainPos = val - 1
			}
			continue
		}

		val, err := strconv.Atoi(arg)
		if err != nil {
			return nil, fmt.Errorf("bad keyword number: %s", arg)
		}
		switch inputType {
		case 0:
			k.SingularPos = val - 1
		case 1:
			k.PluralPos = val - 1
		default:
			return nil, fmt.Errorf("bad keyword number: %s", arg)
		}
		inputType++
	}

	return k, nil
}

var bilingualKeywordTemplates = []Keyword{
	{Name: "Get", IDToken: etype.Singular, SingularPos: 0, PluralPos: -1, ContextPos: -1, DomainPos: -1},
	{Name: "DGet", IDToken: etype.Singular, DomainPos: 0, SingularPos: 1, PluralPos: -1, ContextPos: -1},
	{Name: "NGet", IDToken: etype.Singular, SingularPos: 0, PluralPos: 1, ContextPos: -1, DomainPos: -1},
	{Name: "DNGet", IDToken: etype.Singular, DomainPos: 0, SingularPos: 1, PluralPos: 2, ContextPos: -1},
	{Name: "PGet", IDToken: etype.Singular, ContextPos: 0, SingularPos: 1, PluralPos: -1, DomainPos: -1},
	{Name: "DPGet", IDToken: etype.Singular, DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: -1},
	{Name: "NPGet", IDToken: etype.Singular, ContextPos: 0, SingularPos: 1, PluralPos: 2, DomainPos: -1},
	{Name: "DNPGet", IDToken: etype.Singular, DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: 3},
}

var monolingualKeywordTemplates = []Keyword{
	{Name: "Get", IDToken: etype.Key, SingularPos: 0, PluralPos: -1, ContextPos: -1, DomainPos: -1},
	{Name: "DGet", IDToken: etype.Key, DomainPos: 0, SingularPos: 1, PluralPos: -1, ContextPos: -1},
	{Name: "NGet", IDToken: etype.PluralKey, SingularPos: 0, PluralPos: 0, ContextPos: -1, DomainPos: -1},
	{Name: "DNGet", IDToken: etype.PluralKey, DomainPos: 0, SingularPos: 1, PluralPos: 1, ContextPos: -1},
	{Name: "PGet", IDToken: etype.Singular, ContextPos: 0, SingularPos: 1, PluralPos: -1, DomainPos: -1},
	{Name: "DPGet", IDToken: etype.Singular, DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: -1},
	{Name: "NPGet", IDToken: etype.PluralKey, ContextPos: 0, SingularPos: 1, PluralPos: 1, DomainPos: -1},
	{Name: "DNPGet", IDToken: etype.PluralKey, DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: 2},
}

func DefaultKeywords(name string, isMonolingual bool) []*Keyword {
	name = strings.Trim(strings.TrimSpace(name), ".")
	if name == "" {
		name = "T"
	}

	var keywordTemplates []Keyword
	if isMonolingual {
		keywordTemplates = monolingualKeywordTemplates
	} else {
		keywordTemplates = bilingualKeywordTemplates
	}

	keywords := make([]*Keyword, 0, len(keywordTemplates)*4)

	for _, prefix := range []string{"$.", "."} {
		for _, suffix := range []string{"", "f"} {
			for _, tmpl := range keywordTemplates {
				keyword := tmpl
				keyword.Name = prefix + name + "." + tmpl.Name + suffix
				keywords = append(keywords, &keyword)
			}
		}
	}

	return keywords
}

func (k *Keyword) MaxPosition() int {
	start := maxOf(k.SingularPos, -1)
	start = maxOf(start, k.PluralPos)
	start = maxOf(start, k.ContextPos)
	return maxOf(start, k.DomainPos)
}

func maxOf(x, y int) int {
	if x > y {
		return x
	}
	return y
}
