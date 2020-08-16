package dom

// https://dom.spec.whatwg.org/#documenttype
type DocumentType struct {
	name     DOMString
	publicId DOMString
	systemId DOMString
	*NodeFields
}
