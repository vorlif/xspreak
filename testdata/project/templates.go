package main

// xspreak: template
var t = `
{{.T.Get "Hello"}}
`


// xspreak: template
const constTemplate = `


{{.T.NGet "Dog" "Dogs" 2 "2" "b"}}


`

// xspreak: template
const multiline = "{{   .T.Get `Multiline String\nwith\n  newlines` }}"