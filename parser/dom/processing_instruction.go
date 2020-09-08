package dom

import "browser/parser/webidl"

// https://dom.spec.whatwg.org/#processinginstruction
type ProcessingInstruction struct {
	target webidl.DOMString
	*CharacterData
}
