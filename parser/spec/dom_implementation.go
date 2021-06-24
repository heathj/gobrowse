package spec

//https:domspec.whatwg.org/#domimplementation
type DOMImplementation struct {
}

func (d *DOMImplementation) CreateDocumentType(qualifiedName, publicID, systemID string) *DocumentType {
	return nil
}
func (d *DOMImplementation) CreateDocument(namespace, qualifiedName string, docType DocumentType) *XMLDocument {
	return nil
}
func (d *DOMImplementation) CreateHTMLDocument(title string) *Document {
	return nil
}

func (d *DOMImplementation) HasFeature() bool { return true }
