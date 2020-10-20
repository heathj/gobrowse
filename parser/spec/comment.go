package spec

import "browser/parser/webidl"

// Comment is https:domspec.whatwg.org/#interface-comment
type Comment struct {
	*CharacterData
}

// NewComment returns a comment node with its Data section filled.
func NewComment(data webidl.DOMString) *Comment {
	return &Comment{
		CharacterData: &CharacterData{
			Data:   data,
			Length: len(data),
		},
	}
}

// NewCommentDefault constructor with the default data section of empty string.
func NewCommentDefault() *Comment {
	return NewComment("")
}
