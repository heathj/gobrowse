package spec

import "github.com/heathj/gobrowse/parser/webidl"

// https:domspec.whatwg.org/#processinginstruction
type ProcessingInstruction struct {
	Target webidl.DOMString
	*CharacterData
}
