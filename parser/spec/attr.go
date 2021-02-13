package spec

import (
	"github.com/heathj/gobrowse/parser/webidl"
)

// Attr is https:domspec.whatwg.org/#attr
type Attr struct {
	Namespace                      Namespace
	Prefix, LocalName, Name, Value webidl.DOMString
	OwnerElement                   *Node
	Specified                      bool
}

func NewAttr(k string, attr *Attr, oe *Node) *Attr {
	return &Attr{
		Namespace:    attr.Namespace,
		Prefix:       attr.Prefix,
		LocalName:    attr.LocalName,
		Name:         attr.LocalName,
		Value:        attr.Value,
		OwnerElement: oe,
	}
}

func (a *Attr) QualifiedName() string {
	if a.Prefix == "" {
		return string(a.LocalName)
	}

	return string(a.Prefix) + ":" + string(a.LocalName)
}
