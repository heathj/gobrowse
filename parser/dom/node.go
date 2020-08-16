package dom

import "browser/parser/webidl"

type nodeType uint16

const (
	elementNode nodeType = iota + 1
	attrNode
	textNode
	cdataSectionNode
	processingInstructionNode
	commentNode
	documentNode
	documentTypeNode
	documentFragmentNode
)

type documentPosition uint16

const (
	disconnected           documentPosition = 0x01
	preceding              documentPosition = 0x02
	following              documentPosition = 0x04
	contain                documentPosition = 0x08
	containedBy            documentPosition = 0x10
	implementationSpecific documentPosition = 0x20
)

// https://dom.spec.whatwg.org/#node
type Node interface {
	getRootNode(o GetRootNodeOptions) *Node
	hasChildNodes() bool
	normalize()
	cloneNodeDef() *Node
	cloneNode(deep bool) *Node
	isEqualNode(on *Node) bool
	isSameNode(on *Node) bool
	compareDocumentPosition(on *Node) documentPosition
	contains(on *Node) bool
	lookupPrefix(namespace webidl.DOMString) webidl.DOMString
	lookupNamespaceURI(prefix webidl.DOMString) webidl.DOMString
	isDefaultNamespace() bool
	insertBefore(on *Node, child *Node) *Node
	appendChild(on *Node) *Node
	replaceChild(on *Node, child *Node) *Node
	removeChild(child *Node) *Node
}
type NodeFields struct {
	nodeType        nodeType
	nodeName        webidl.DOMString
	baseURI         webidl.USVString
	isConnected     bool
	ownerDocument   *Document
	parentNode      *Node
	parentElement   *Element
	childNodes      NodeList
	firstChild      *Node
	lastChild       *Node
	previousSibling *Node
	nextSibling     *Node
	nodeValue       webidl.DOMString
	textContent     webidl.DOMString

	*EventTarget
}

func (n *NodeFields) getRootNode(o GetRootNodeOptions) *Node {
	return nil
}
func (n *NodeFields) hasChildNodes() bool {
	return len(n.childNodes) > 0
}
func (n *NodeFields) normalize() {}
func (n *NodeFields) cloneNodeDef() *Node {
	return n.cloneNode(false)
}
func (n *NodeFields) cloneNode(deep bool) *Node {
	return nil
}
func (n *NodeFields) isEqualNode(on *Node) bool                                   { return false }
func (n *NodeFields) isSameNode(on *Node) bool                                    { return false }
func (n *NodeFields) compareDocumentPosition(on *Node) documentPosition           { return disconnected }
func (n *NodeFields) contains(on *Node) bool                                      { return false }
func (n *NodeFields) lookupPrefix(namespace webidl.DOMString) webidl.DOMString    { return "" }
func (n *NodeFields) lookupNamespaceURI(prefix webidl.DOMString) webidl.DOMString { return "" }
func (n *NodeFields) isDefaultNamespace() bool                                    { return false }
func (n *NodeFields) insertBefore(on *Node, child *Node) *Node                    { return nil }
func (n *NodeFields) appendChild(on *Node) *Node                                  { return nil }
func (n *NodeFields) replaceChild(on *Node, child *Node) *Node                    { return nil }
func (n *NodeFields) removeChild(child *Node) *Node                               { return nil }
