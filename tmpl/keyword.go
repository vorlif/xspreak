package tmpl

import (
	"fmt"
	"strconv"
	"strings"
)

type Keyword struct {
	Name        string
	SingularPos int
	PluralPos   int
	ContextPos  int
	DomainPos   int
}

func ParseKeywords(spec string) (*Keyword, error) {
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
		SingularPos: 0,
		PluralPos:   -1,
		ContextPos:  -1,
		DomainPos:   -1,
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

func DefaultKeywords(prefix string) []*Keyword {
	prefix = strings.TrimSuffix(strings.TrimSpace(prefix), ".")
	if prefix == "" {
		prefix = ".T"
	}

	return []*Keyword{
		{Name: prefix + ".Get", SingularPos: 0, PluralPos: -1, ContextPos: -1, DomainPos: -1},
		{Name: prefix + ".Getf", SingularPos: 0, PluralPos: -1, ContextPos: -1, DomainPos: -1},

		{Name: prefix + ".DGet", DomainPos: 0, SingularPos: 1, PluralPos: -1, ContextPos: -1},
		{Name: prefix + ".DGetf", DomainPos: 0, SingularPos: 1, PluralPos: -1, ContextPos: -1},

		{Name: prefix + ".NGet", SingularPos: 0, PluralPos: 1, ContextPos: -1, DomainPos: -1},
		{Name: prefix + ".NGetf", SingularPos: 0, PluralPos: 1, ContextPos: -1, DomainPos: -1},

		{Name: prefix + ".DNGet", DomainPos: 0, SingularPos: 1, PluralPos: 2, ContextPos: -1},
		{Name: prefix + ".DNGetf", DomainPos: 0, SingularPos: 1, PluralPos: 2, ContextPos: -1},

		{Name: prefix + ".PGet", ContextPos: 0, SingularPos: 1, PluralPos: -1, DomainPos: -1},
		{Name: prefix + ".PGetf", ContextPos: 0, SingularPos: 1, PluralPos: -1, DomainPos: -1},

		{Name: prefix + ".DPGet", DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: -1},
		{Name: prefix + ".DPGetf", DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: -1},

		{Name: prefix + ".NPGet", ContextPos: 0, SingularPos: 1, PluralPos: 2, DomainPos: -1},
		{Name: prefix + ".NPGetf", ContextPos: 0, SingularPos: 1, PluralPos: 2, DomainPos: -1},

		{Name: prefix + ".DNPGet", DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: 3},
		{Name: prefix + ".DNPGetf", DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: 3},
	}
}

func (k *Keyword) MaxIndex() int {
	start := max(k.SingularPos, -1)
	start = max(start, k.PluralPos)
	start = max(start, k.ContextPos)
	return max(start, k.DomainPos)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
