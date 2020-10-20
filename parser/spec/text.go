package spec

import "browser/parser/webidl"

// https:domspec.whatwg.org/#text
type Text struct {
	wholeText webidl.DOMString
	*CharacterData
}

func NewText(data webidl.DOMString) *Text {
	return &Text{
		wholeText: data,
		CharacterData: &CharacterData{
			Data:   data,
			Length: len(data),
		}}
}

func (t *Text) splitText(offset uint) *Text {
	return nil
}
