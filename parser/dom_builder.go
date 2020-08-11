package parser

type quirkType uint

const (
	force quirkType = iota
	limited
)

type Document struct {
	docType  documentType
	children []*Element
}

type Element struct {
	elemType elementType
}

type elementType uint

const (
	htmlElement elementType = iota
	tableElement
	tbodyElement
	tfootElement
	theadElement
	trElement
	templateElement
)

type documentType uint

const (
	iframeSrcDoc documentType = iota
)

type DOMBuilder struct {
	quirks                   quirkType
	document                 *Document
	framesetOK               bool
	openElements             []*Element
	headPointer              *Element
	activeFormattingElements []formattingElement
}

type formattingElement uint

const (
	markerFElement formattingElement = iota
	aFElement
	bFElement
	bigFElement
	codeFElement
	emFElement
	fontFElement
	iFElement
	nobrFElement
	sFElement
	smallFElement
	strikeFElement
	strongFElement
	ttFElement
	uFElement
)

func (d *DOMBuilder) SearchOpenElements(e elementType) bool {
	for _, oe := range d.openElements {
		if oe.elemType == e {
			return true
		}
	}
	return false
}

// WriteComment inserts a comment into the DOM.
func (d *DOMBuilder) WriteComment(t *Token) {

}

func (d *DOMBuilder) WriteMarker() {
	d.activeFormattingElements = append(d.activeFormattingElements, markerFElement)
}

// CreateHTMLElement creates an html element whose node document is the Document object.
// Append it to the Document object.
func (d *DOMBuilder) CreateHTMLElement() *Element {
	e := &Element{}
	d.document.children = append(d.document.children, e)
	return e
}

//WriteHTMLElement inserts a foreign element for the token, in the HTML namespace.
func (d *DOMBuilder) WriteHTMLElement(t *Token) {

}

//WriteCharacter adds the given characer to the appropriate place for inserting a node.
func (d *DOMBuilder) WriteCharacter(t *Token) {

}

//WriteLatestHeadElement sets the head element pointer to the newly created head element.
func (d *DOMBuilder) WriteLatestHeadElement() {

}

// PushOpenElements pushes an element to the list of currently open elements being parsed.
func (d *DOMBuilder) PushOpenElements(e *Element) {
	d.openElements = append(d.openElements, e)
}

// PopOpenElements pops an element off the list of open elements
func (d *DOMBuilder) PopOpenElements() {

}

// WriteDocumentType sets the current document type given the curren token.
func (d *DOMBuilder) WriteDocumentType(t *Token) {}

// Quirks sets the quirks value to "force".
func (d *DOMBuilder) Quirks() {
	d.quirks = force
}

// LimitedQuirks sets the quirks value to "limited"
func (d *DOMBuilder) LimitedQuirks() {
	d.quirks = limited
}

// FramesetNotOK sets framesetOK state to false.
func (d *DOMBuilder) FramesetNotOK() {
	d.framesetOK = false
}

// CurrentNode returns the bottommost node from the stack of open elements.
func (d *DOMBuilder) CurrentNode() *Element {
	return d.openElements[len(d.openElements)-1]
}

// HeadPointer returns the current element that is set as the head pointer.
func (d *DOMBuilder) HeadPointer() *Element {
	return d.headPointer
}

func (d *DOMBuilder) FormPointer() *Element {
	return d.formPointer
}

// Document builds the document from the current builder state.
func (d *DOMBuilder) Document() *Document {
	return d.document
}
