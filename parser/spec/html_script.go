package spec

type HTMLScript struct {
	Src                                                             string
	ScriptElementType, CrossOrigin, Text, Integrity, ReferrerPolicy string
	NoModule, Async, DeferScript, NonBlocking, AlreadyStated        bool
	ParserDocument                                                  *HTMLDocument
}
