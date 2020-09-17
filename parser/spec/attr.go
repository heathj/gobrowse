package spec

import "browser/parser/webidl"

// Attr is https:domspec.whatwg.org/#attr
type Attr struct {
	NamespaceURI webidl.DOMString
	Prefix       webidl.DOMString
	LocalName    webidl.DOMString
	Name         webidl.DOMString
	Value        webidl.DOMString
	OwnerElement *Element
	Specified    bool
}
