package etype

type Token string

const (
	None      Token = ""
	Singular  Token = "Singular"
	Plural    Token = "Plural"
	Domain    Token = "Domain"
	Context   Token = "Context"
	Message   Token = "Message"
	Key       Token = "Key"
	PluralKey Token = "PluralKey"
)

func IsMessageID(tok Token) bool {
	return tok == Singular || tok == Key || tok == PluralKey
}

var StringExtractNames = map[string]Token{
	"MsgID":     Singular,
	"Singular":  Singular,
	"Plural":    Plural,
	"Domain":    Domain,
	"Context":   Context,
	"Key":       Key,
	"PluralKey": PluralKey,
}
