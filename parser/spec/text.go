package spec

import "browser/parser/webidl"

// https:domspec.whatwg.org/#text
type Text struct {
	wholeText webidl.DOMString
	*CharacterData
}

func (t *Text) splitText(offset uint) *Text {
	return nil
}
