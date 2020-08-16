package dom

import "browser/parser/webidl"

//https://dom.spec.whatwg.org/#interface-comment
type Comment struct {
	*CharacterData
	*NodeFields
}

// NewComment returns a comment node with its Data section filled.
func NewComment(data webidl.DOMString) *Comment {
	return &Comment{
		CharacterData: &CharacterData{
			Data:   data,
			Length: uint(len(data)),
		},
	}
}

// NewComment constructor with the default data section of empty string.
func NewCommentDefault() *Comment {
	return NewComment("")
}
