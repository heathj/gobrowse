package spec

import "browser/parser/webidl"

// DocumentType is https:domspec.whatwg.org/#documenttype
type DocumentType struct {
	Name     webidl.DOMString
	PublicID webidl.DOMString
	SystemID webidl.DOMString
}
