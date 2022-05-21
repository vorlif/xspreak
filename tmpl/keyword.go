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

var keywordTemplates = []Keyword{
	{Name: "Get", SingularPos: 0, PluralPos: -1, ContextPos: -1, DomainPos: -1},
	{Name: "DGet", DomainPos: 0, SingularPos: 1, PluralPos: -1, ContextPos: -1},
	{Name: "NGet", SingularPos: 0, PluralPos: 1, ContextPos: -1, DomainPos: -1},
	{Name: "DNGet", DomainPos: 0, SingularPos: 1, PluralPos: 2, ContextPos: -1},
	{Name: "PGet", ContextPos: 0, SingularPos: 1, PluralPos: -1, DomainPos: -1},
	{Name: "DPGet", DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: -1},
	{Name: "NPGet", ContextPos: 0, SingularPos: 1, PluralPos: 2, DomainPos: -1},
	{Name: "DNPGet", DomainPos: 0, ContextPos: 1, SingularPos: 2, PluralPos: 3},
}

func DefaultKeywords(name string) []*Keyword {
	name = strings.Trim(strings.TrimSpace(name), ".")
	if name == "" {
		name = "T"
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
