xspreak is the command line program for extracting strings for the [spreak library](https://github.com/vorlif/spreak).

# xspreak ![Test status](https://github.com/vorlif/xspreak/workflows/Test/badge.svg) [![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

xspreak automatically extracts strings that use a string alias from
the [`localize` package](https://pkg.go.dev/github.com/vorlif/spreak/localize).
The extracted strings are stored in a `.pot` or `.json` file and can then be easily translated.

The extracted strings can then be passed to a [Localizer](https://pkg.go.dev/github.com/vorlif/spreak#Localizer)
or a [KeyLocalizer](https://pkg.go.dev/github.com/vorlif/spreak#KeyLocalizer) which returns the matching translation.

Example:

```go
package main

import (
	"fmt"

	"golang.org/x/text/language"

	"github.com/vorlif/spreak"
	"github.com/vorlif/spreak/localize"
)

// This string is extracted because the type is localize.Singular
var ApplicationName localize.Singular = "Beautiful app"

func main() {
	bundle, err := spreak.NewBundle(
		spreak.WithSourceLanguage(language.English),
		spreak.WithDomainPath(spreak.NoDomain, "../locale"),
		spreak.WithLanguage(language.German, language.Spanish, language.Chinese),
	)
	if err != nil {
		panic(err)
	}

	t := spreak.NewLocalizer(bundle, language.Spanish)

	// Message lookup of the extracted string
	fmt.Println(t.Get(ApplicationName))
	// Output:
	// Hermosa app
}
```

## Requirements

* Your project must be a go module (must have a `go.mod` and `go.sum`)
* The dependencies of your project must be installed `go mod tidy`
* xspreak searches all files for strings to extract. This can take a lot of memory or CPU time for larger projects.

## How to install

Download a [pre-built binary from the releases](https://github.com/vorlif/xspreak/releases/latest) or create it from
source:

```bash
go install github.com/vorlif/xspreak@latest
```

Tests installation with:

```bash
xspreak --help
```

## What can be extracted?

### spreak functions calls

All function calls of spreak translation functions, where a string is passed, are extracted.

```go
t.Get("Hello world")
t.Nget("Hello world", "Hello worlds")
// ...
```

### Global variables and constants

Global variables and constants are extracted if the type is `localize.Singular` or `localize.MsgID`.
Thereby `localize.Singular` and `localize.MsgID` are always equivalent and can be used synonymously.

```go
package main

import "github.com/vorlif/spreak/localize"

const Weekday localize.Singular = "weekday"

var ApplicationName localize.Singular = "app"
```

### Local variables

Local variables are extracted if the type is `localize.Singular` or `localize.MsgID`.

```go
package main

import "github.com/vorlif/spreak/localize"

func init() {
	holiday := localize.Singular("Christmas")
}
```

### Variable assignments

Assignments to variables are extracted if the type is `localize.Singular` or `localize.MsgID`.

```go
package main

import "github.com/vorlif/spreak/localize"

var ApplicationName = "app"

func init() {
	var holiday localize.Singular

	holiday = "Mother's Day"

	ApplicationName = "App for you"
}
```

### Argument of function calls

Function calls to **global functions** are extracted if the parameter type is from the `localize` package.
The parameters of a function are grouped together to form a message.
Thus, a message can be created with singular, plural, a context and a domain.

```go
package main

import "github.com/vorlif/spreak/localize"

func noop(name localize.Singular, plural localize.Plural, ctx localize.Context) {}

func init() {
	// Extracted as a message with singular, plural and a context
	noop("I have %d car", "I have %d cars", "cars")
}
```

### Return values of functions

Return values of functions are extracted if the parameter type is from the `localize` package.
The parameters of a function are grouped together to form a message.
Thus, a message can be created with singular, plural, a context and a domain.
Named return values are currently not supported.

```go
package main

import "github.com/vorlif/spreak/localize"

func noop() (localize.Singular, localize.Plural, localize.Context) {
	// Extracted as a message with singular, plural and a context
	return "I have %d car", "I have %d cars", "cars"
}
```

### Attributes at struct initialization

Struct initializations are extracted if the struct was **defined globally** and
the attribute type *comes from the `localize` package*.
The attributes of a struct are grouped together to create a message.
Thus, a message can be created with singular, plural, a context and a domain.

```go
package main

import "github.com/vorlif/spreak/localize"

type MyMessage struct {
	// Defined as singular and plural
	Text   localize.Singular
	Plural localize.Plural
	Tmp    string
}

func main() {
	msg := &MyMessage{
		// Extracted as a message with singular and plural
		Text:   "Hello planet",
		Plural: "Hello planets",

		// not extracted - type string
		Tmp: "tmp",
	}
}
```

### Values from an array initialization

Arrays are extracted if the type is `localize.Singular` or a struct that contains parameter
types from the `localize` package.

```go
package main

import "github.com/vorlif/spreak/localize"

var weekdays = []localize.MsgID{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

type MyMessage struct {
	Text   localize.Singular
	Plural localize.Plural
	Tmp    string
}

func main() {
	animals := []MyMessage{
		{Text: "%d dog", Plural: "%d dogs"},
		{Text: "%d cat", Plural: "%d cat"},
		{Text: "%d horse", Plural: "%d horses"},
	}
}
```

### Values from a map initialization

During map initialization keys and values are extracted,
if the type is `localize.Singular` or a struct that contains parameter types from the `localize` package.

```go
package main

import "github.com/vorlif/spreak/localize"

var weekdays = map[localize.MsgID]int{
	"Monday":  1,
	"Tuesday": 2,
}

var reverseWeekdays = map[int]localize.Singular{
	1: "Monday",
	2: "Tuesday",
}

```

### Error texts

Strings can be extracted from `errors.New` if xspreak is called with the `-e` option.

```go
package main

import "errors"

var ErrInvalidAnimal = errors.New("this is not a valid animal")

```

### Comments

Comments can be left for translators.
These are extracted, stored in the `.pot` file and displayed to the translator.
For JSON files the comments are ignored.

```go
package main

import "github.com/vorlif/spreak/localize"

// TRANSLATORS: This comment is automatically extracted by xspreak
// and can be used to leave useful hints for the translators.
//
// This comment is not extracted because a blank line was inserted above it.
const InvalidName localize.Singular = "The name has an invalid format"
```

### Exclude from extraction

Strings can be ignored.

```go
package main

import "github.com/vorlif/spreak/localize"

// xspreak: ignore
const MagicName localize.Singular = ".%$($§($(%"
```

### Templates (Experimental)

With `-t` a template directory can be specified.

* `-t "path/to/templates/*.html"`: Scans all HTML files in the `templates` directory
* `-t "path/to/templates/**/*.html"` Scans all HTML files in the `templates` directory
  and in subdirectories of the templates directory.

With `--template-prefix` you can specify a prefix for the template function.

For example, with `--template-prefix "T"` the following function calls are extracted

```text
{{.T.Get "Hello world"}}
{{.T.NGet "I see a planet" "I see planets" 2}}
```

Two messages are extracted here:

* Message 1
    * Singular `Hello world`
* Message 2
    * Singular `I see a planet`
    * Plural `I see planets`

Instead of `--template-prefix` you can also use `-k` to define your own keywords.
The definition follows
the [xgettext notation](https://www.gnu.org/software/gettext/manual/html_node/xgettext-Invocation.html),
but only applies to templates.

The default is: `.T.Get .T.Getf .T.DGet:1d,2 .T.DGetf:1d,2 .T.NGet:1,2 .T.NGetf:1,2 .T.DNGet:1d,2,3 .T.DNGetf:1d,2,3
.T.PGet:1c,2 .T.PGetf:1c,2 .T.DPGet:1d,2c,3 .T.DPGetf:1d,2c,3 .T.NPGet:1c,2,3 .T.NPGetf:1c,2,3 .T.DNPGet:1d,2c,3,4
.T.DNPGetf:1d,2c,3,4`

Inline templates must be marked with `xspreak: template`:

```go
package main

// xspreak: template
const tmpl = `{{.T.Get "Hello"}}`
```

There is also a [detailed example](https://github.com/vorlif/spreak/tree/main/examples/features/httptempl) how to use
spreak with templates and your own keywords.

### Variables tracing

For `localize.Singular` and `localize.MsgID`, variable tracing is supported for simple cases.
The following example extracts two strings. One string with "Yes" and one string with "No, better a beer".

```go
package main

import "github.com/vorlif/spreak/localize"

func WantCoffee() localize.Singular {
	answer := "Yes"

	switch time.Now().Weekday() {
	case time.Friday, time.Saturday, time.Sunday:
		answer = "No, better a beer"
	}

	return answer
}
```

Whether tracing works for a particular use case should be verified separately for each use case.

### Using monolingual format (e.g. Key-Value)

To use monolingual format, the following changes must be made.

1. Instead of `localize.Singular` and `localize.MsgID` use `localize.Key`.
2. If a key also has a plural form, `localize.PluralKey` must be used.
3. When extracting templates `--template-use-kv` must be added
4. If a separate key is defined for templates, the position of singular and plural must be identical for plural:
   ```shell
   # Example
   xspreak -f json -D ./ -p locale/ -k "i18n.TrN:1,1" --template-use-kv -t "templates/*.html"
   ```

All of the above functions also apply to `localize.Key` and `localize.PluralKey`.

5. Use [KeyLocalizer](https://pkg.go.dev/github.com/vorlif/spreak#KeyLocalizer) instead of Localizer in the code

### Supported export formats

1. `po`/`pot` (Default) `xspreak ...`
2. `json`: `xspreak -f json ...`
