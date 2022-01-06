package spec

import (
	"fmt"
	"sort"
	"strings"
)

func (c *NodeList) ContainsElementInSpecificScopeExcept(target string, list ...string) bool {
	for i := len(*c) - 1; i >= 0; i-- {
		entry := (*c)[i]
		if target == entry.NodeName {
			return true
		}

		allNotMatch := false
		for _, name := range list {
			if entry.NodeName == name {
				allNotMatch = true
				break
			}
		}
		if !allNotMatch {
			return false
		}
	}

	return false
}

var elementInScopeList = []string{
	"applet",
	"caption",
	"html",
	"table",
	"td",
	"th",
	"marquee",
	"object",
	"template",
	"mi",
	"mo",
	"mn",
	"ms",
	"mtext",
	"annotation-xml",
	"foreignObject",
	"desc",
	"title",
}
var listItemScopeList = append(elementInScopeList, "ol", "ul")
var buttonItemScopeList = append(elementInScopeList, "button")

func (c *NodeList) ContainsElementInSpecificScope(target string, list ...string) bool {
	for i := len(*c) - 1; i >= 0; i-- {
		entry := (*c)[i]
		if target == entry.NodeName {
			return true
		}

		for _, name := range list {
			if entry.NodeName == name {
				return false
			}
		}
	}

	return false
}

func (c *NodeList) ContainsElementsInScope(elems ...string) bool {
	for _, elem := range elems {
		if c.ContainsElementInScope(elem) {
			return true
		}
	}

	return false
}

func (c *NodeList) ContainsElementInScope(target string) bool {
	return c.ContainsElementInSpecificScope(target, elementInScopeList...)
}

func (c *NodeList) ContainsElementInListItemScope(target string) bool {
	return c.ContainsElementInSpecificScope(target, listItemScopeList...)
}

func (c *NodeList) ContainsElementInButtonScope(target string) bool {
	return c.ContainsElementInSpecificScope(target, buttonItemScopeList...)
}

func (c *NodeList) ContainsElementInTableScope(target string) bool {
	return c.ContainsElementInSpecificScope(target, "html", "table", "template")
}

func (c *NodeList) ContainsElementInSelectScope(target string) bool {
	return c.ContainsElementInSpecificScopeExcept(target, "optgroup", "option")
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
	Disconnected      DocumentPosition = 0x01
	Preceding         DocumentPosition = 0x02
	Following         DocumentPosition = 0x04
	Contain           DocumentPosition = 0x08
	ContainedBy       DocumentPosition = 0x10
	Implementationfic DocumentPosition = 0x20
)

var ScopeMarker = &Node{
	NodeType: ScopeMarkerNode,
	NodeName: "marker",
}

// NewComment returns a comment node with its Data section filled.
func NewComment(data string, od *Node) *Node {
	return &Node{
		NodeType:      CommentNode,
		OwnerDocument: od,
		Comment: &Comment{
			CharacterData: &CharacterData{
				Data:   data,
				Length: len(data),
			},
		}}
}

func NewHTMLDocumentNode() *HTMLDocument {
	return &HTMLDocument{
		Node: &Node{
			NodeType: DocumentNode,
			Document: &Document{Type: "html"},
		},
	}
}

func NewTextNode(od *Node, text string) *Node {
	return &Node{
		NodeType:      TextNode,
		OwnerDocument: od,
		Text: &Text{
			CharacterData: &CharacterData{
				Data: text,
			},
		},
	}
}

func NewDocTypeNode(name, pub, sys string) *Node {
	return &Node{
		NodeType: DocumentTypeNode,
		NodeName: name,
		DocumentType: &DocumentType{
			Name:     name,
			PublicID: pub,
			SystemID: sys,
		},
	}
}

func NewDOMElement(od *Node, name string, namespace Namespace, optionals ...string) *Node {
	// handle custom events? not sure how that will work since that is functionality of the HTML
	// not the DOM  might need to create a shared package or something so I don't get
	// circular deps
	var prefix string
	if len(optionals) >= 1 {
		prefix = optionals[0]
	}
	n := &Node{
		NodeType:      ElementNode,
		NodeName:      name,
		OwnerDocument: od,
		Element: &Element{
			NamespaceURI: namespace,
			Prefix:       prefix,
			LocalName:    name,
			Attributes:   NewNamedNodeMap(map[string]*Attr{}, nil),
			HTMLElement:  NewHTMLElement(name),
		},
	}

	n.Attributes.AssociatedElement = n
	return n
}

// https://dom.whatwg.org/#node
type Node struct {
	NodeType                                                        NodeType
	NodeName                                                        string
	BaseURI                                                         string
	IsConnected                                                     bool
	OwnerDocument                                                   *Node
	ParentNode, FirstChild, LastChild, PreviousSibling, NextSibling *Node
	ParentElement                                                   *Element
	ChildNodes                                                      NodeList
	NodeValue, TextContent                                          string

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

func serializeNodeType(node *Node, ident int) string {
	switch node.NodeType {
	case ElementNode:
		e := "<"
		switch node.Element.NamespaceURI {
		case Svgns:
			e += "svg "
		case Mathmlns:
			e += "math "
		}
		e += string(node.NodeName)
		if node.Attributes != nil && len(node.Attributes.Attrs) != 0 {
			e += ">"
			keys := make([]string, 0, len(node.Attributes.Attrs))
			for name := range node.Attributes.Attrs {
				keys = append(keys, string(name))
			}
			sort.Strings(keys)
			spaces := "| "
			for i := 1; i < ident; i++ {
				spaces += "  "
			}
			for _, name := range keys {
				attr := node.Attributes.Attrs[name]
				var ns string
				switch attr.Namespace {
				case Xmlnsns:
					ns = "xmlns "
				case Xmlns:
					ns = "xml "
				case Xlinkns:
					ns = "xlink "
				case Svgns:
					ns = "svg "
				case Mathmlns:
					ns = "math "
				}
				e += "\n" + spaces + ns + name + "=\"" + string(attr.Value) + "\""
			}
		} else {
			e += ">"
		}
		return e
	case TextNode:
		return "\"" + string(node.Text.Data) + "\""
	case CommentNode:
		return "<!-- " + string(node.Comment.Data) + " -->"
	case DocumentTypeNode:
		d := "<!DOCTYPE " + string(node.DocumentType.Name)
		if len(node.DocumentType.PublicID) == 0 && len(node.DocumentType.SystemID) == 0 {
			return d
		}
		if (len(node.DocumentType.PublicID) != 0 && string(node.DocumentType.PublicID) != "MISSING") ||
			(len(node.DocumentType.SystemID) != 0 && string(node.DocumentType.SystemID) != "MISSING") {
			if node.DocumentType.PublicID == "MISSING" {
				d += " \"" + "\""
			} else {
				d += " \"" + string(node.DocumentType.PublicID) + "\""
			}

			if node.DocumentType.SystemID == "MISSING" {
				d += " \"" + "\""
			} else {
				d += " \"" + string(node.DocumentType.SystemID) + "\""
			}
		}

		d += ">"
		return d
	case DocumentNode:
		return "#document"
	case ProcessingInstructionNode:
		return "<?" + string(node.ProcessingInstruction.CharacterData.Data) + ">"
	default:
		fmt.Printf("Error serializing node: %+v\n", node.NodeType)
		return ""
	}

}

func (node *Node) serialize(ident int) string {
	ser := serializeNodeType(node, ident+1) + "\n"
	if node.NodeType != DocumentNode {
		spaces := "| "
		for i := 1; i < ident; i++ {
			spaces += "  "
		}
		ser = spaces + ser
	}
	for _, child := range node.ChildNodes {
		ser += child.serialize(ident + 1)
	}

	return ser
}

func (node *Node) String() string {
	return strings.TrimRight(node.serialize(0), "\n")
} //

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
	copy := &Node{}
	if n.NodeType == ElementNode {
		copy = NewDOMElement(n, n.NodeName, n.Element.NamespaceURI, n.Element.Prefix)
		attrs := make(map[string]*Attr)
		for k, v := range n.Attributes.Attrs {
			attrs[string(k)] = v
		}
		copy.Attributes = NewNamedNodeMap(attrs, copy)
	} else {
		copy.NodeType = n.NodeType
		switch n.NodeType {
		case DocumentNode:
			copy.Document = &Document{}
			copy.InputEncoding = n.InputEncoding
			copy.ContentType = n.ContentType
			copy.URL = n.URL
			// origin
			// type
			copy.CompatMode = n.CompatMode
		case DocumentTypeNode:
			copy.DocumentType = &DocumentType{}
			copy.DocumentType.Name = n.DocumentType.Name
			copy.PublicID = n.PublicID
			copy.SystemID = n.SystemID
		case AttrNode:
			copy.Attr = &Attr{}
			copy.Attr.Namespace = n.Attr.Namespace
			copy.Attr.Prefix = n.Attr.Prefix
			copy.Attr.LocalName = n.Attr.LocalName
			copy.Attr.Value = n.Attr.Value
		case TextNode:
			copy.Text = NewText(n.Text.Data)
		case CommentNode:
			copy.Comment = NewComment(n.Comment.Data, n.OwnerDocument).Comment
		case ProcessingInstructionNode:
			copy.ProcessingInstruction = &ProcessingInstruction{}
			copy.ProcessingInstruction.Target = n.ProcessingInstruction.Target
			copy.ProcessingInstruction.Data = n.ProcessingInstruction.Data
		}
	}

	if copy.NodeType == DocumentNode {
		copy.OwnerDocument = copy
		n.OwnerDocument = copy //? I don't think this is right
	} else {
		copy.OwnerDocument = n.OwnerDocument
	}

	if deep {
		for _, child := range n.ChildNodes {
			copy.AppendChild(child.CloneNode(true))
		}
	}
	copy.ParentNode = n.ParentNode
	copy.FirstChild = n.FirstChild
	copy.PreviousSibling = n.PreviousSibling
	copy.NextSibling = n.NextSibling

	return copy
}

// https://dom.whatwg.org/#concept-node-equals
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

func (n *Node) IsSameNode(on *Node) bool                          { return false }
func (n *Node) CompareDocumentPosition(on *Node) DocumentPosition { return Disconnected }
func (n *Node) Contains(on *Node) bool                            { return false }
func (n *Node) LookupPrefix(namespace string) string              { return "" }
func (n *Node) LookupNamespaceURI(prefix string) string           { return "" }
func (n *Node) IsDefaultNamespace() bool                          { return false }
func (n *Node) InsertBefore(on, child *Node) *Node {

	//old := n.getRoot().String()
	for i := range n.ChildNodes {
		if n.ChildNodes[i] == child {
			n.ChildNodes = append(n.ChildNodes[:i+1], n.ChildNodes[i:]...)
			n.ChildNodes[i] = on
			on.ParentNode = n
			on.NextSibling = child
			if i == 0 {
				n.FirstChild = on
			} else {
				on.PreviousSibling = n.ChildNodes[i-1]
			}
		}
	}

	//new := n.getRoot().String()
	//PrintDiff(old, new, "InsertBefore")
	return on
}

// didn't really follow the steps here because they seem complicated :/
// https://dom.whatwg.org/#concept-node-append
func (n *Node) AppendChild(on *Node) *Node {
	//old := n.getRoot().String()
	if n.LastChild != nil {
		on.PreviousSibling = n.LastChild
		n.LastChild.NextSibling = on
	}
	on.ParentNode = n
	n.LastChild = on
	n.ChildNodes = append(n.ChildNodes, on)

	//new := n.getRoot().String()
	//PrintDiff(old, new, "AppendChild")
	return on
}
func (n *Node) ReplaceChild(on, child *Node) *Node { return nil }

// TODO: not to yet, for some reason remove is like 50 steps. I'll come back to it
func (n *Node) RemoveChild(child *Node) *Node {
	//old := n.getRoot().String()
	node := n.ChildNodes.Remove(n.ChildNodes.Contains(child))
	if n.LastChild != nil {
		if len(n.ChildNodes) == 0 {
			n.LastChild = nil
		} else {
			n.LastChild = n.ChildNodes[len(n.ChildNodes)-1]
		}
	}

	//new := n.getRoot().String()
	//PrintDiff(old, new, "RemoveChild")
	return node
}

func (n *Node) getRoot() *Node {
	var prev *Node
	for i := n; i != nil; i = i.ParentNode {
		prev = i
	}

	return prev
}

func PrintDiff(a, b, method string) {
	//not very helpful logs
	/*if a == b {
		return
	}
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(a, b, true)
	logrus.WithField("method", method).Debugf("[TREE]: %s\n\n", dmp.DiffPrettyText(diffs))
	*/
}
