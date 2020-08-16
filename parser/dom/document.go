package dom

import "browser/parser/webidl"

// https://dom.spec.whatwg.org/#interface-document
type Document struct {
	implementation  DOMImplementation
	URL             webidl.USVString
	documentURI     webidl.USVString
	compatMode      webidl.DOMString
	characterSet    webidl.DOMString
	charset         webidl.DOMString
	inputEncoding   webidl.DOMString
	contentType     webidl.DOMString
	doctype         DocumentType
	documentElement *Element

	*NodeFields
}

func (d *Document) getElementsByTagName(qualifiedName webidl.DOMString) HTMLCollection { return nil }
func (d *Document) getElementsByTagNameNS(namespace, localName webidl.DOMString) HTMLCollection {
	return nil
}
func (d *Document) getElementsByClassName(classNames webidl.DOMString) HTMLCollection { return nil }
func (d *Document) createElement(localName, options webidl.DOMString) *Element        { return nil }
func (d *Document) createElementWithOpts(localName webidl.DOMString, options ElementCreationOptions) *Element {
	return nil
}
func (d *Document) createElementNS(namespace, qualifiedName, options webidl.DOMString) *Element {
	return nil
}
func (d *Document) createElementNSWithOpts(namespace, qualifiedName webidl.DOMString, options ElementCreationOptions) *Element {
	return nil
}
func (d *Document) createDocumentFragment() *DocumentFragment              { return nil }
func (d *Document) createTextNode(data webidl.DOMString) *Text             { return nil }
func (d *Document) createCDATASection(data webidl.DOMString) *CDATASection { return nil }
func (d *Document) createComment(data webidl.DOMString) *Comment           { return nil }
func (d *Document) createProcessingInstruction(target, data webidl.DOMString) *ProcessingInstruction {
	return nil
}
func (d *Document) importNode(node *Node, deep bool) *Node                            { return nil }
func (d *Document) importNodeDefault(node *Node) *Node                                { return nil }
func (d *Document) adoptNode(node *Node) *Node                                        { return nil }
func (d *Document) createAttribute(localName webidl.DOMString) *Attr                  { return nil }
func (d *Document) createAttributeNS(namespace, qualifiedName webidl.DOMString) *Attr { return nil }
func (d *Document) createEvent(ifc webidl.DOMString) *Event                           { return nil }
func (d *Document) createRange() *Range                                               { return nil }
func (d *Document) createNodeIterator(root *Node, whatToShow uint, filter NodeFilter) *NodeIterator {
	return nil
}
func (d *Document) createTreeWalker(root *Node, whatToShow uint, filter NodeFilter) *TreeWalker {
	return nil
}
