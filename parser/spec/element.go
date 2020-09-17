package spec

import (
	"browser/parser/webidl"
)

// https:domspec.whatwg.org/#htmlcollection
type HTMLCollection []*Element

// Element is an individual HTML element that gets added to the spec\.
// https:domspec.whatwg.org/#interface-element
type Element struct {
	NamespaceURI, Prefix, LocalName, TagName, Id, ClassName, Slot webidl.DOMString
	ClassList                                                     DOMTokenList
	Attributes                                                    *NamedNodeMap

	*HTMLElement
}

func (e *Element) HasAttributes() bool                                                { return false }
func (e *Element) GetAttributeNames() []webidl.DOMString                              { return nil }
func (e *Element) GetAttribute(qualifiedName webidl.DOMString) webidl.DOMString       { return "" }
func (e *Element) GetAttributeNS(namespace, value webidl.DOMString) webidl.DOMString  { return "" }
func (e *Element) SetAttribute(qualifiedName, value webidl.DOMString)                 {}
func (e *Element) SetAttributeNS(namespace, qualifiedName, value webidl.DOMString)    {}
func (e *Element) RemoveAttribute(qualifiedName webidl.DOMString)                     {}
func (e *Element) RemoveAttributeNS(namespace, localName webidl.DOMString)            {}
func (e *Element) ToggleAttribute(qualifiedName webidl.DOMString, force ...bool) bool { return false }
func (e *Element) HasAttribute(qualifiedName webidl.DOMString) bool                   { return false }
func (e *Element) HasAttributeNS(namespace, localName webidl.DOMString) bool          { return false }
func (e *Element) GetAttributeNode(qualifiedName webidl.DOMString) *Attr              { return nil }
func (e *Element) GetAttributeNodeNS(namespace, localName webidl.DOMString) *Attr     { return nil }
func (e *Element) SetAttributeNode(attr *Attr) *Attr                                  { return nil }
func (e *Element) SetAttributeNodeNS(attr *Attr) *Attr                                { return nil }
func (e *Element) RemoveAttributeNode(attr *Attr) *Attr                               { return nil }
func (e *Element) AttachShadow(init *ShadowRootInit) *ShadowRoot                      { return nil }
func (e *Element) Closest(selectors webidl.DOMString) *Element                        { return nil }
func (e *Element) Matches(selectors webidl.DOMString) bool                            { return false }
func (e *Element) WebkitMatchesSelector(selectors webidl.DOMString) bool              { return false }
func (e *Element) GetElementsByTagName(qualifiedName webidl.DOMString) HTMLCollection { return nil }
func (e *Element) GetElementsByTagNameNS(namespace, localName webidl.DOMString) HTMLCollection {
	return nil
}
func (e *Element) GetElementsByClassName(qualifiedName webidl.DOMString) HTMLCollection    { return nil }
func (e *Element) InsertAdjacentElement(where webidl.DOMString, element *Element) *Element { return nil }
func (e *Element) InsertAdjacentTExt(where, data webidl.DOMString)                         {}

type ElementType uint

const (
	HtmlElement ElementType = iota
	TableElement
	TbodyElement
	TfootElement
	TheadElement
	TrElement
	TemplateElement
	DocumentElement
)
