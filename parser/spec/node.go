package spec

import (
	"browser/parser/webidl"
)

func Contains(n *Node, h *[]*Node) int {
	for i := range *h {
		if n == (*h)[i] {
			return i
		}
	}
	return -1
}

func Remove(i int, h *[]*Node) {
	if i == -1 {
		return
	}
	*h = append((*h)[:i], (*h)[i+1:]...)
}

func Pop(h *[]*Node) {
	*h = (*h)[:len((*h))-1]
}

func Push(h *[]*Node, n *Node) {
	*h = append(*h, n)
}

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
	ScopeMarkerNode
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

var ScopeMarker = &Node{
	NodeType: ScopeMarkerNode,
}

func NewHTMLDocumentNode() *HTMLDocument {
	return &HTMLDocument{
		Node: &Node{
			NodeType: DocumentNode,
			Document: &Document{},
		},
	}
}

func NewTextNode(od *Node, text string) *Node {
	return &Node{
		NodeType:      TextNode,
		OwnerDocument: od,
		Text: &Text{
			CharacterData: &CharacterData{
				Data: webidl.DOMString(text),
			},
		},
	}
}

func NewDocTypeNode(name, pub, sys string) *Node {
	return &Node{
		NodeType: DocumentTypeNode,
		NodeName: webidl.DOMString(name),
		DocumentType: &DocumentType{
			Name:     webidl.DOMString(name),
			PublicID: webidl.DOMString(pub),
			SystemID: webidl.DOMString(sys),
		},
	}
}

func NewDOMElement(od *Node, name, namespace webidl.DOMString, optionals ...webidl.DOMString) *Node {
	// handle custom events? not sure how that will work since that is functionality of the HTML
	// spec not the DOM spec. might need to create a shared package or something so I don't get
	// circular deps
	var prefix webidl.DOMString
	if len(optionals) >= 1 {
		prefix = optionals[0]
	}
	return &Node{
		NodeType:      ElementNode,
		NodeName:      name,
		OwnerDocument: od,
		Element: &Element{
			NamespaceURI: namespace,
			Prefix:       prefix,
			LocalName:    name,
		},
	}
}

// https://dom.spec.whatwg.org/#node
type Node struct {
	NodeType                                                        NodeType
	NodeName                                                        webidl.DOMString
	BaseURI                                                         webidl.USVString
	IsConnected                                                     bool
	OwnerDocument                                                   *Node
	ParentNode, FirstChild, LastChild, PreviousSibling, NextSibling *Node
	ParentElement                                                   *Element
	ChildNodes                                                      NodeList
	NodeValue, TextContent                                          webidl.DOMString

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
		if n.DocumentType.Name != on.DocumentType.Name || n.PublicID != on.PublicID || n.SystemID != on.SystemID {
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

// didn't really follow the steps here because they seem complicated :/
// https://dom.spec.whatwg.org/#concept-node-append
func (n *Node) AppendChild(on *Node) *Node {
	if n.LastChild != nil {
		on.PreviousSibling = n.LastChild
		n.LastChild.NextSibling = on
	}
	on.ParentNode = n
	n.LastChild = on
	n.ChildNodes = append(n.ChildNodes, on)
	return on
}
func (n *Node) ReplaceChild(on, child *Node) *Node { return nil }
func (n *Node) RemoveChild(child *Node) *Node      { return nil }
