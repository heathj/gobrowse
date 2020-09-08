package dom

import "browser/parser/webidl"

// https://dom.spec.whatwg.org/#dictdef-elementcreationoptions
type ElementCreationOptions map[elementCreationKeys]webidl.DOMString
type elementCreationKeys uint

const (
	is elementCreationKeys = iota
)
