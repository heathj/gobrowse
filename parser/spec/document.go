package spec

// Document is https:domspec.whatwg.org/#interface-document
type Document struct {
	Implementation                                                DOMImplementation
	URL, DocumentURI                                              string
	CompatMode, CharacterSet, Charset, InputEncoding, ContentType string
	Doctype                                                       *Node
	DocumentElement                                               *Element

	Origin string
	Mode   string
	Type   string
}

// GetElementsByTagName is https:domspec.whatwg.org/#dom-document-getelementsbytagname
func (d *Document) GetElementsByTagName(qualifiedName string) HTMLCollection { return nil }

// GetElementsByTagNameNS is https:domspec.whatwg.org/#dom-document-getelementsbytagnamens
func (d *Document) GetElementsByTagNameNS(namespace, localName string) HTMLCollection {
	return nil
}

// GetElementsByClassName is
func (d *Document) GetElementsByClassName(classNames string) HTMLCollection { return nil }

// CreateElement is
func (d *Document) CreateElement(localName string, options ...string) *Element {
	return nil
}

// CreateElementWithOpts is
func (d *Document) CreateElementWithOpts(localName string, options ElementCreationOptions) *Element {
	return nil
}

// CreateElementNS is
func (d *Document) CreateElementNS(namespace, qualifiedName, options string) *Element {
	return nil
}

// CreateElementNSWithOpts is
func (d *Document) CreateElementNSWithOpts(namespace, qualifiedName string, options ElementCreationOptions) *Element {
	return nil
}
func (d *Document) CreateDocumentFragment() *DocumentFragment    { return nil }
func (d *Document) CreateTextNode(data string) *Text             { return nil }
func (d *Document) CreateCDATASection(data string) *CDATASection { return nil }
func (d *Document) CreateComment(data string) *Comment           { return nil }
func (d *Document) CreateProcessingInstruction(target, data string) *ProcessingInstruction {
	return nil
}
func (d *Document) ImportNode(node *Node, deep bool) *Node                  { return nil }
func (d *Document) ImportNodeDefault(node *Node) *Node                      { return nil }
func (d *Document) AdoptNode(node *Node) *Node                              { return nil }
func (d *Document) CreateAttribute(localName string) *Attr                  { return nil }
func (d *Document) CreateAttributeNS(namespace, qualifiedName string) *Attr { return nil }
func (d *Document) CreateEvent(ifc string) *Event                           { return nil }
func (d *Document) CreateRange() *Range                                     { return nil }
func (d *Document) CreateNodeIterator(root *Node, whatToShow uint, filter NodeFilter) *NodeIterator {
	return nil
}
func (d *Document) CreateTreeWalker(root *Node, whatToShow uint, filter NodeFilter) *TreeWalker {
	return nil
}
