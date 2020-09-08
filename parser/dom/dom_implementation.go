package dom

import "browser/parser/webidl"

//https://dom.spec.whatwg.org/#domimplementation
type DOMImplementation struct {
}

func (d *DOMImplementation) CreateDocumentType(qualifiedName, publicID, systemID webidl.DOMString) *DocumentType {
	return nil
}
func (d *DOMImplementation) CreateDocument(namespace, qualifiedName webidl.DOMString, docType DocumentType) *XMLDocument {
	return nil
}
func (d *DOMImplementation) CreateHTMLDocument(title webidl.DOMString) *Document {
	return nil
}

func (d *DOMImplementation) HasFeature() bool { return true }
