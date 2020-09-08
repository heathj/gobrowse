package dom

import (
	"browser/parser/webidl"
)

type NodeType uint16

const (
	ElementNode NodeType = iota + 1
	AttrNode
	TextNode
	CDATASectionNode
	ProcessingInstructionNode
	CommentNode
	DocumentNode
	DocumentTypeNode
	DocumentFragmentNode
)

type DocumentPosition uint16

const (
	Disconnected           DocumentPosition = 0x01
	Preceding              DocumentPosition = 0x02
	Following              DocumentPosition = 0x04
	Contain                DocumentPosition = 0x08
	ContainedBy            DocumentPosition = 0x10
	ImplementationSpecific DocumentPosition = 0x20
)

// https://dom.spec.whatwg.org/#node
type Node struct {
	NodeType        NodeType
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

	// Node types
	*Element
	*Attr
	*Text
	*CDATASection
	*ProcessingInstruction
	*Comment
	*Document
	*DocumentType
	*DocumentFragment
}

func (n *Node) GetRootNode(o GetRootNodeOptions) *Node {
	return nil
}
func (n *Node) HasChildNodes() bool {
	return len(n.ChildNodes) > 0
}
func (n *Node) Normalize() {}
func (n *Node) CloneNodeDef() *Node {
	return n.CloneNode(false)
}
func (n *Node) CloneNode(deep bool) *Node {
	return nil
}

// https://dom.spec.whatwg.org/#concept-node-equals
func (n *Node) IsEqualNode(on *Node) bool {
	// if on.(*Node).NodeType != n.NodeType {
	// 	return false
	// }

	switch n.NodeType {
	case DocumentTypeNode:
		if n.Name != on.Name || n.PublicID != on.PublicID || n.SystemID != on.SystemID {
			return false
		}
	case ElementNode:
	case AttrNode:
	case ProcessingInstructionNode:
	case TextNode, CommentNode:
	default:
	}

	return true
}

func (n *Node) IsSameNode(on *Node) bool                                    { return false }
func (n *Node) CompareDocumentPosition(on *Node) DocumentPosition           { return Disconnected }
func (n *Node) Contains(on *Node) bool                                      { return false }
func (n *Node) LookupPrefix(namespace webidl.DOMString) webidl.DOMString    { return "" }
func (n *Node) LookupNamespaceURI(prefix webidl.DOMString) webidl.DOMString { return "" }
func (n *Node) IsDefaultNamespace() bool                                    { return false }
func (n *Node) InsertBefore(on, child *Node) *Node                          { return nil }
func (n *Node) AppendChild(on *Node) *Node                                  { return nil }
func (n *Node) ReplaceChild(on, child *Node) *Node                          { return nil }
func (n *Node) RemoveChild(child *Node) *Node                               { return nil }
