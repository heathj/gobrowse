package parser

// https://dom.spec.whatwg.org/#htmlcollection
type HTMLCollection []Element

type quirkType uint

const (
	force quirkType = iota
	limited
)

//https://dom.spec.whatwg.org/#domimplementation
type DOMImplementation struct {
}

func (d *DOMImplementation) createDocumentType() {}
func (d *DOMImplementation) createDocument()     {}
func (d *DOMImplementation) createHTMLDocument() {}

// https://dom.spec.whatwg.org/#documenttype
type DocumentType struct {
	name     DOMString
	publicId DOMString
	systemId DOMString
	Node
}

// https://dom.spec.whatwg.org/#dictdef-elementcreationoptions
type ElementCreationOptions map[elementCreationKeys]DOMString
type elementCreationKeys uint

const (
	is elementCreationKeys = iota
)

// https://dom.spec.whatwg.org/#documentfragment
type DocumentFragment struct {
	Node
}

// https://dom.spec.whatwg.org/#characterdata
type CharacterData struct {
	data   DOMString
	length uint

	Node
}

func (c *CharacterData) substringData(offset, count uint) DOMString     { return "" }
func (c *CharacterData) appendData(data DOMString)                      {}
func (c *CharacterData) insertData(offset uint, data DOMString)         {}
func (c *CharacterData) deleteData(offset, count uint)                  {}
func (c *CharacterData) replaceData(offset, count uint, data DOMString) {}

// https://dom.spec.whatwg.org/#text
type Text struct {
	wholeText DOMString
	CharacterData
}

func (t *Text) splitText(offset uint) *Text {
	return nil
}

//https://dom.spec.whatwg.org/#interface-cdatasection
type CDATASection struct {
	Text
}

//https://dom.spec.whatwg.org/#interface-comment
type Comment struct {
	CharacterData
}

// https://dom.spec.whatwg.org/#processinginstruction
type ProcessingInstruction struct {
	target DOMString
	CharacterData
}

// https://dom.spec.whatwg.org/#attr
type Attr struct {
	namespaceURI DOMString
	prefix       DOMString
	localName    DOMString
	name         DOMString
	value        DOMString
	ownerElement *Element
	specified    bool
	Node
}
type NodeFilter struct{}

// https://dom.spec.whatwg.org/#nodeiterator
type NodeIterator struct {
	root                       *Node
	referenceNode              *Node
	pointerBeforeReferenceNode bool
	whatToShow                 uint
	filter                     NodeFilter
}

func (n *NodeIterator) nextNode() *Node     { return nil }
func (n *NodeIterator) previousNode() *Node { return nil }
func (n *NodeIterator) detach()             {}

type eventPhase uint

const (
	noneEventPhase eventPhase = iota
	capturingPhase
	atTargetPhase
	bubblingPhase
)

type DOMHighResTimeStamp uint64
type Event struct {
	eventType        DOMString
	target           EventTarget
	srcElement       EventTarget
	currentTarget    EventTarget
	eventPhase       eventPhase
	cancelBubble     bool
	bubbles          bool
	cancelable       bool
	returnValue      bool
	defaultPrevented bool
	composed         bool
	isTrusted        bool
	timeStamp        DOMHighResTimeStamp
}

func (e *Event) composedPath() []EventTarget                    { return nil }
func (e *Event) stopPropagation()                               {}
func (e *Event) stopImmediatePropagation()                      {}
func (e *Event) preventDefault()                                {}
func (e *Event) initEvent(eventType DOMString, options ...bool) {}

// https://dom.spec.whatwg.org/#abstractrange
type AbstractRange struct {
	startContainer *Node
	startOffset    uint
	endContainer   *Node
	endOffSet      uint
	collapsed      bool
}

type howRange uint

const (
	starToStart howRange = iota
	starToEnd
	endToEnd
	endToStart
)

//https://dom.spec.whatwg.org/#range
type Range struct {
	commonAncestorContainer *Node

	AbstractRange
}

func (r *Range) setStart(node *Node, offset uint)                            {}
func (r *Range) setEnd(node *Node, offset uint)                              {}
func (r *Range) setStartBefore(node *Node)                                   {}
func (r *Range) setStartAfter(node *Node)                                    {}
func (r *Range) setEndBefore(node *Node)                                     {}
func (r *Range) setEndAfter(node *Node)                                      {}
func (r *Range) collapse(toStart bool)                                       {}
func (r *Range) collapseDef()                                                {}
func (r *Range) selectNode(node *Node)                                       {}
func (r *Range) selectNodeContents(node *Node)                               {}
func (r *Range) compareBoundaryPoints(how howRange, sourceRange Range) int16 { return 0 }
func (r *Range) deleteContents()                                             {}
func (r *Range) extractContents() *DocumentFragment                          { return nil }
func (r *Range) cloneContents() *DocumentFragment                            { return nil }
func (r *Range) insertNode(node *Node)                                       {}
func (r *Range) surroundContents(newparent *Node)                            {}
func (r *Range) cloneRange() *Range                                          { return nil }
func (r *Range) detach()                                                     {}
func (r *Range) isPointInRange(node *Node, offset uint) bool                 { return false }
func (r *Range) comparePoint(node *Node, offset uint) int16                  { return 0 }
func (r *Range) intersectsNode(node *Node) bool                              { return false }

// https://dom.spec.whatwg.org/#treewalker
type TreeWalker struct {
	root        *Node
	whatToShow  uint
	filter      NodeFilter
	currentNode *Node
}

func (t *TreeWalker) parentNode() *Node      { return nil }
func (t *TreeWalker) firstChild() *Node      { return nil }
func (t *TreeWalker) lastChild() *Node       { return nil }
func (t *TreeWalker) previousSibling() *Node { return nil }
func (t *TreeWalker) nextSibling() *Node     { return nil }
func (t *TreeWalker) previousNode() *Node    { return nil }
func (t *TreeWalker) nextNode() *Node        { return nil }

// Document holds the current document that is being parsed.
// https://dom.spec.whatwg.org/#document
type Document struct {
	implementation  DOMImplementation
	URL             USVString
	documentURI     USVString
	compatMode      DOMString
	characterSet    DOMString
	charset         DOMString
	inputEncoding   DOMString
	contentType     DOMString
	doctype         DocumentType
	documentElement Element

	Node
}

func (d *Document) getElementsByTagName(qualifiedName DOMString) HTMLCollection          { return nil }
func (d *Document) getElementsByTagNameNS(namespace, localName DOMString) HTMLCollection { return nil }
func (d *Document) getElementsByClassName(classNames DOMString) HTMLCollection           { return nil }
func (d *Document) createElement(localName, options DOMString) *Element                  { return nil }
func (d *Document) createElementWithOpts(localName DOMString, options ElementCreationOptions) *Element {
	return nil
}
func (d *Document) createElementNS(namespace, qualifiedName, options DOMString) *Element { return nil }
func (d *Document) createElementNSWithOpts(namespace, qualifiedName DOMString, options ElementCreationOptions) *Element {
	return nil
}
func (d *Document) createDocumentFragment() *DocumentFragment       { return nil }
func (d *Document) createTextNode(data DOMString) *Text             { return nil }
func (d *Document) createCDATASection(data DOMString) *CDATASection { return nil }
func (d *Document) createComment(data DOMString) *Comment           { return nil }
func (d *Document) createProcessingInstruction(target, data DOMString) *ProcessingInstruction {
	return nil
}
func (d *Document) importNode(node *Node, deep bool) *Node                     { return nil }
func (d *Document) importNodeDefault(node *Node) *Node                         { return nil }
func (d *Document) adoptNode(node *Node) *Node                                 { return nil }
func (d *Document) createAttribute(localName DOMString) *Attr                  { return nil }
func (d *Document) createAttributeNS(namespace, qualifiedName DOMString) *Attr { return nil }
func (d *Document) createEvent(ifc DOMString) *Event                           { return nil }
func (d *Document) createRange() *Range                                        { return nil }
func (d *Document) createNodeIterator(root *Node, whatToShow uint, filter NodeFilter) *NodeIterator {
	return nil
}
func (d *Document) createTreeWalker(root *Node, whatToShow uint, filter NodeFilter) *TreeWalker {
	return nil
}
