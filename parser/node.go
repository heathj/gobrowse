package parser

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

// https://heycam.github.io/webidl/#idl-DOMString
type DOMString string

// https://heycam.github.io/webidl/#idl-USVString
type USVString string

//https://dom.spec.whatwg.org/#nodelist
type NodeList []Node

// https://dom.spec.whatwg.org/#dictdef-getrootnodeoptions
type GetRootNodeOptions struct {
	composed bool
}

// https://dom.spec.whatwg.org/#node
type Node struct {
	nodeType        nodeType
	nodeName        DOMString
	baseURI         USVString
	isConnected     bool
	ownerDocument   *Document
	parentNode      *Node
	parentElement   *Element
	childNodes      NodeList
	firstChild      *Node
	lastChild       *Node
	previousSibling *Node
	nextSibling     *Node
	nodeValue       DOMString
	textContent     DOMString

	EventTarget
}

func (n *Node) getRootNode(o GetRootNodeOptions) *Node {
	return nil
}
func (n *Node) hasChildNodes() bool {
	return len(n.childNodes) > 0
}
func (n *Node) normalize() {}
func (n *Node) cloneNodeDef() *Node {
	return n.cloneNode(false)
}
func (n *Node) cloneNode(deep bool) *Node {
	return nil
}
func (n *Node) isEqualNode(on *Node) bool                         { return false }
func (n *Node) isSameNode(on *Node) bool                          { return false }
func (n *Node) compareDocumentPosition(on *Node) documentPosition { return disconnected }
func (n *Node) contains(on *Node) bool                            { return false }
func (n *Node) lookupPrefix(namespace DOMString) DOMString        { return "" }
func (n *Node) lookupNamespaceURI(prefix DOMString) DOMString     { return "" }
func (n *Node) isDefaultNamespace() bool                          { return false }
func (n *Node) insertBefore(on *Node, child *Node) *Node          { return nil }
func (n *Node) appendChild(on *Node) *Node                        { return nil }
func (n *Node) replaceChild(on *Node, child *Node) *Node          { return nil }
func (n *Node) removeChild(child *Node) *Node                     { return nil }
