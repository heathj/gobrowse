package spec

import (
	"browser/parser/webidl"
	"strings"
)

func NewNamedNodeMap(attrs map[string]*Attr, oe *Node) *NamedNodeMap {
	a := make(map[webidl.DOMString]*Attr, len(attrs))
	for k, v := range attrs {
		a[webidl.DOMString(k)] = NewAttr(k, v, oe)
	}
	return &NamedNodeMap{
		Length:            len(a),
		Attrs:             a,
		AssociatedElement: oe,
	}
}

type NamedNodeMap struct {
	Length            int
	Attrs             map[webidl.DOMString]*Attr
	AssociatedElement *Node
}

func (n *NamedNodeMap) GetNamedItem(qn webidl.DOMString) *Attr {
	return n.getAttributeByName(qn)

}

func (n *NamedNodeMap) getAttributeByName(qn webidl.DOMString) *Attr {
	if n.AssociatedElement.OwnerDocument != nil &&
		n.AssociatedElement.Element.NamespaceURI == Htmlns &&
		n.AssociatedElement.OwnerDocument.NodeType == DocumentNode &&
		n.AssociatedElement.OwnerDocument.Type == "html" {
		qn = webidl.DOMString(strings.ToLower(string(qn)))
	}

	if v, ok := n.Attrs[qn]; ok {
		return v
	}

	return nil
}

func (n *NamedNodeMap) getAttributeByNSLocalName(ns Namespace, ln webidl.DOMString) *Attr {
	if v, ok := n.Attrs[ln]; ok {
		if v.Namespace == ns {
			return v
		}
	}

	return nil
}

func (n *NamedNodeMap) SetNamedItem(s *Attr) *Attr {
	s.OwnerElement = n.AssociatedElement
	if s == nil {
		return nil
	}

	oldAttr := n.getAttributeByNSLocalName(s.Namespace, s.LocalName)
	if oldAttr == nil {
		n.Attrs[s.LocalName] = s
		return s
	}
	if oldAttr == s {
		return s
	}

	return oldAttr
}

func (n *NamedNodeMap) GetNamedItemNS(ns Namespace, ln webidl.DOMString) *Attr {
	return n.getAttributeByNSLocalName(ns, ln)
}
func (n *NamedNodeMap) SetNamedItemNS(attr *Attr) *Attr { return nil }
func (n *NamedNodeMap) RemoveNamedItem() *Attr          { return nil }
func (n *NamedNodeMap) RemoveNamedItemNS() *Attr        { return nil }
