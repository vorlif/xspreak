package encoder

import (
	"regexp"

	"github.com/vorlif/xspreak/result"
)

// Recognizes not all cases, but most. - See Unit tests.
var reGoStringFormat = regexp.MustCompile(`%([#+\-*0.])?(\[\d])?(([1-9])\.([1-9])|([1-9])|([1-9])\.|\.([1-9]))?[xsvTtbcdoOqXUeEfFgGp]`)

type Encoder interface {
	Encode(issues []result.Issue) error
}
