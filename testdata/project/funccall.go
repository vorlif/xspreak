package main

import (
	"golang.org/x/text/language"

	"github.com/vorlif/testdata/sub"

	"github.com/vorlif/spreak"
	sp "github.com/vorlif/spreak"
	alias "github.com/vorlif/spreak/localize"
)

func noop(sing alias.MsgID, plural alias.Plural, context alias.Context, domain alias.Domain) {}
func noParamNames(alias.MsgID, alias.Plural)                                                 {}
func variadicFunc(a alias.Singular, vars ...alias.Singular)                                  {}
func multiNamesFunc(a, b alias.MsgID)                                                        {}
func ctxAndMsg(c alias.Context, msgid alias.Singular)                                        {}

func GenericFunc[V int64 | float64](log alias.Singular, i V) V {
	return i
}

type methodStruct struct{}

func (methodStruct) Method(alias.Singular) {}

type genericMethodStruct[T any] struct{}

func (genericMethodStruct[T]) Method(alias.Singular) {}

func outerFuncDef() {
	f := func(msgid alias.Singular, plural alias.Plural, context alias.Context, domain alias.Domain) {}

	variadicFunc("pre-variadic", "variadic-a", "variadic-b")
	noParamNames("no-param-msgid", "no-param-plural")
	multiNamesFunc("multi-names-a", "multi-names-b")

	// not extracted
	f("f-msgid", "f-plural", "f-context", "f-domain")

	// extracted
	noop("noop-msgid", "noop-plural", "noop-context", "noop-domain")
	sub.Func("submsgid", "subplural")
	_ = GenericFunc[int64]("generic-call", 5)
}

// TRANSLATORS: this is not extracted
func localizerCall(loc *sp.Localizer) {
	// TRANSLATORS: this is extracted
	loc.Getf("localizer func call")

	initBacktrace := "init backtrace"
	loc.Get(initBacktrace)

	var assignBacktrace string
	assignBacktrace = "assign backtrace"
	loc.Get(assignBacktrace)

	ctxAndMsg(constCtx, "constCtxMsg")
}

func builtInFunctions() {
	bundle, err := spreak.NewBundle(
		spreak.WithDefaultDomain(spreak.NoDomain),
		spreak.WithDomainPath(spreak.NoDomain, "./"),
		spreak.WithLanguage(language.English),
	)
	if err != nil {
		panic(err)
	}

	inlineFunc := func(inlineParam alias.Singular) {}

	inlineFunc("inline function")

	t := spreak.NewLocalizer(bundle, "en")
	// TRANSLATORS: Test
	// multiline
	t.Getf("msgid")
	t.NGetf("msgid-n", "pluralid-n", 10, 10)
	t.DGetf("domain-d", "msgid-d")
	t.DNGetf("domain-dn", "msgid-dn", "pluralid-dn", 10)
	t.PGetf("context-pg", "msgid-pg")
	t.NPGetf("context-np", "msgid-np", "pluralid-np", 10)
	t.DPGetf("domain-dp", "context-dp", "singular-dp")
	t.DNPGetf("domain-dnp", "context-dnp", "msgid-dnp", "pluralid-dnp", 10)
}

func methodCall() {
	(methodStruct{}).Method("struct-method-call")
	(genericMethodStruct[string]{}).Method("generic-struct-method-call")
}
