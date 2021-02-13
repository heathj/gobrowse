package spec

import "github.com/heathj/gobrowse/parser/webidl"

// DocumentType is https:domspec.whatwg.org/#documenttype
type DocumentType struct {
	Name     webidl.DOMString
	PublicID webidl.DOMString
	SystemID webidl.DOMString
}
