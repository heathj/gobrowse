package dom

import "browser/parser/webidl"

// Attr is https://dom.spec.whatwg.org/#attr
type Attr struct {
	namespaceURI webidl.DOMString
	prefix       webidl.DOMString
	localName    webidl.DOMString
	name         webidl.DOMString
	value        webidl.DOMString
	ownerElement *Element
	specified    bool
}
