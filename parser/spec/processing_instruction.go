package spec

import "browser/parser/webidl"

// https:domspec.whatwg.org/#processinginstruction
type ProcessingInstruction struct {
	target webidl.DOMString
	*CharacterData
}
