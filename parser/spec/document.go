package spec

import "browser/parser/webidl"

// Document is https:domspec.whatwg.org/#interface-document
type Document struct {
	Implementation                                                DOMImplementation
	URL, DocumentURI                                              webidl.USVString
	CompatMode, CharacterSet, Charset, InputEncoding, ContentType webidl.DOMString
	Doctype                                                       *Node
	DocumentElement                                               *Element
}

// GetElementsByTagName is https:domspec.whatwg.org/#dom-document-getelementsbytagname
func (d *Document) GetElementsByTagName(qualifiedName webidl.DOMString) HTMLCollection { return nil }

// GetElementsByTagNameNS is https:domspec.whatwg.org/#dom-document-getelementsbytagnamens
func (d *Document) GetElementsByTagNameNS(namespace, localName webidl.DOMString) HTMLCollection {
	return nil
}

// GetElementsByClassName is
func (d *Document) GetElementsByClassName(classNames webidl.DOMString) HTMLCollection { return nil }

// CreateElement is
func (d *Document) CreateElement(localName webidl.DOMString, options ...webidl.DOMString) *Element {
	return nil
}

// CreateElementWithOpts is
func (d *Document) CreateElementWithOpts(localName webidl.DOMString, options ElementCreationOptions) *Element {
	return nil
}

// CreateElementNS is
func (d *Document) CreateElementNS(namespace, qualifiedName, options webidl.DOMString) *Element {
	return nil
}

// CreateElementNSWithOpts is
func (d *Document) CreateElementNSWithOpts(namespace, qualifiedName webidl.DOMString, options ElementCreationOptions) *Element {
	return nil
}
func (d *Document) CreateDocumentFragment() *DocumentFragment              { return nil }
func (d *Document) CreateTextNode(data webidl.DOMString) *Text             { return nil }
func (d *Document) CreateCDATASection(data webidl.DOMString) *CDATASection { return nil }
func (d *Document) CreateComment(data webidl.DOMString) *Comment           { return nil }
func (d *Document) CreateProcessingInstruction(target, data webidl.DOMString) *ProcessingInstruction {
	return nil
}
func (d *Document) ImportNode(node *Node, deep bool) *Node                            { return nil }
func (d *Document) ImportNodeDefault(node *Node) *Node                                { return nil }
func (d *Document) AdoptNode(node *Node) *Node                                        { return nil }
func (d *Document) CreateAttribute(localName webidl.DOMString) *Attr                  { return nil }
func (d *Document) CreateAttributeNS(namespace, qualifiedName webidl.DOMString) *Attr { return nil }
func (d *Document) CreateEvent(ifc webidl.DOMString) *Event                           { return nil }
func (d *Document) CreateRange() *Range                                               { return nil }
func (d *Document) CreateNodeIterator(root *Node, whatToShow uint, filter NodeFilter) *NodeIterator {
	return nil
}
func (d *Document) CreateTreeWalker(root *Node, whatToShow uint, filter NodeFilter) *TreeWalker {
	return nil
}

func (d *Document) String() string {
	return ""
}
