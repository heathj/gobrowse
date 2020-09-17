package spec

import "browser/parser/webidl"

// https:domspec.whatwg.org/#dictdef-elementcreationoptions
type ElementCreationOptions map[elementCreationKeys]webidl.DOMString
type elementCreationKeys uint

const (
	is elementCreationKeys = iota
)
