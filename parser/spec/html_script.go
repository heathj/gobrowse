package spec

import "github.com/heathj/gobrowse/parser/webidl"

type HTMLScript struct {
	Src                                                             webidl.USVString
	ScriptElementType, CrossOrigin, Text, Integrity, ReferrerPolicy webidl.DOMString
	NoModule, Async, DeferScript, NonBlocking, AlreadyStated        bool
	ParserDocument                                                  *HTMLDocument
}
