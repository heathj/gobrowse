package dom

import "browser/parser/webidl"

//https://dom.spec.whatwg.org/#domimplementation
type DOMImplementation struct {
}

func (d *DOMImplementation) createDocumentType(qualifiedName, publicID, systemID webidl.DOMString) *DocumentType {
	return nil
}
func (d *DOMImplementation) createDocument(namespace, qualifiedName webidl.DOMString, docType DocumentType) *XMLDocument {
	return nil
}
func (d *DOMImplementation) createHTMLDocument(title webidl.DOMString) *Document {
	return nil
}

func (d *DOMImplementation) hasFeature() bool { return true }
