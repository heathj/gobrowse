package dom

import "browser/parser/webidl"

// DocumentType is https://dom.spec.whatwg.org/#documenttype
type DocumentType struct {
	Name     webidl.DOMString
	PublicID webidl.DOMString
	SystemID webidl.DOMString
}
