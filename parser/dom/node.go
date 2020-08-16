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
	GetRootNode(o GetRootNodeOptions) *Node
	HasChildNodes() bool
	Normalize()
	CloneNodeDef() *Node
	CloneNode(deep bool) *Node
	IsEqualNode(on *Node) bool
	IsSameNode(on *Node) bool
	CompareDocumentPosition(on *Node) documentPosition
	Contains(on *Node) bool
	LookupPrefix(namespace webidl.DOMString) webidl.DOMString
	LookupNamespaceURI(prefix webidl.DOMString) webidl.DOMString
	IsDefaultNamespace() bool
	InsertBefore(on *Node, child *Node) *Node
	AppendChild(on *Node) *Node
	ReplaceChild(on *Node, child *Node) *Node
	RemoveChild(child *Node) *Node
}
type NodeFields struct {
	NodeType        nodeType
	NodeName        webidl.DOMString
	BaseURI         webidl.USVString
	IsConnected     bool
	OwnerDocument   *Document
	ParentNode      *Node
	ParentElement   *Element
	ChildNodes      NodeList
	FirstChild      *Node
	LastChild       *Node
	PreviousSibling *Node
	NextSibling     *Node
	NodeValue       webidl.DOMString
	TextContent     webidl.DOMString

	*EventTarget
}

func (n *NodeFields) GetRootNode(o GetRootNodeOptions) *Node {
	return nil
}
func (n *NodeFields) HasChildNodes() bool {
	return len(n.childNodes) > 0
}
func (n *NodeFields) Normalize() {}
func (n *NodeFields) CloneNodeDef() *Node {
	return n.cloneNode(false)
}
func (n *NodeFields) CloneNode(deep bool) *Node {
	return nil
}
func (n *NodeFields) IsEqualNode(on *Node) bool                                   { return false }
func (n *NodeFields) IsSameNode(on *Node) bool                                    { return false }
func (n *NodeFields) CompareDocumentPosition(on *Node) documentPosition           { return disconnected }
func (n *NodeFields) Contains(on *Node) bool                                      { return false }
func (n *NodeFields) LookupPrefix(namespace webidl.DOMString) webidl.DOMString    { return "" }
func (n *NodeFields) LookupNamespaceURI(prefix webidl.DOMString) webidl.DOMString { return "" }
func (n *NodeFields) IsDefaultNamespace() bool                                    { return false }
func (n *NodeFields) InsertBefore(on *Node, child *Node) *Node                    { return nil }
func (n *NodeFields) AppendChild(on *Node) *Node                                  { return nil }
func (n *NodeFields) ReplaceChild(on *Node, child *Node) *Node                    { return nil }
func (n *NodeFields) RemoveChild(child *Node) *Node                               { return nil }
