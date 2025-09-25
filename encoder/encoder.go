package encoder

import (
	"regexp"

	"github.com/vorlif/xspreak/extract"
)

// Regexp that matches if the string contains a string formatting verb
// like %s, %d, %f, etc.
// Recognizes not all cases but most. - See Unit tests.
var reGoStringFormat = regexp.MustCompile(`%([#+\-*0.])?(\[\d])?(([1-9])\.([1-9])|([1-9])|([1-9])\.|\.([1-9]))?[xsvTtbcdoOqXUeEfFgGp]`)

type Encoder interface {
	Encode(issues []extract.Issue) error
}
