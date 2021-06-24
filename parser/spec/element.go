package spec

type Namespace uint

const (
	Htmlns Namespace = iota
	Mathmlns
	Svgns
	Xlinkns
	Xmlns
	Xmlnsns
)

// https:domspec.whatwg.org/#htmlcollection
type HTMLCollection []*Element

// Element is an individual HTML element that gets added to the spec\.
// https:domspec.whatwg.org/#interface-element
type Element struct {
	NamespaceURI                           Namespace
	Prefix, LocalName, Id, ClassName, Slot string
	ClassList                              DOMTokenList
	Attributes                             *NamedNodeMap

	*HTMLElement
}

func (e *Element) HasAttributes() bool                                      { return false }
func (e *Element) GetAttributeNames() []string                              { return nil }
func (e *Element) GetAttribute(qualifiedName string) string                 { return "" }
func (e *Element) GetAttributeNS(namespace, value string) string            { return "" }
func (e *Element) SetAttribute(qualifiedName, value string)                 {}
func (e *Element) SetAttributeNS(namespace, qualifiedName, value string)    {}
func (e *Element) RemoveAttribute(qualifiedName string)                     {}
func (e *Element) RemoveAttributeNS(namespace, localName string)            {}
func (e *Element) ToggleAttribute(qualifiedName string, force ...bool) bool { return false }
func (e *Element) HasAttribute(qualifiedName string) bool                   { return false }
func (e *Element) HasAttributeNS(namespace, localName string) bool          { return false }
func (e *Element) GetAttributeNode(qualifiedName string) *Attr              { return nil }
func (e *Element) GetAttributeNodeNS(namespace, localName string) *Attr     { return nil }
func (e *Element) SetAttributeNode(attr *Attr) *Attr                        { return nil }
func (e *Element) SetAttributeNodeNS(attr *Attr) *Attr                      { return nil }
func (e *Element) RemoveAttributeNode(attr *Attr) *Attr                     { return nil }
func (e *Element) AttachShadow(init *ShadowRootInit) *ShadowRoot            { return nil }
func (e *Element) Closest(selectors string) *Element                        { return nil }
func (e *Element) Matches(selectors string) bool                            { return false }
func (e *Element) WebkitMatchesSelector(selectors string) bool              { return false }
func (e *Element) GetElementsByTagName(qualifiedName string) HTMLCollection { return nil }
func (e *Element) GetElementsByTagNameNS(namespace, localName string) HTMLCollection {
	return nil
}
func (e *Element) GetElementsByClassName(qualifiedName string) HTMLCollection { return nil }
func (e *Element) InsertAdjacentElement(where string, element *Element) *Element {
	return nil
}
func (e *Element) InsertAdjacentTExt(where, data string) {}

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
