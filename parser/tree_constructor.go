package parser

import (
	"browser/parser/spec"
	"browser/parser/webidl"
	"fmt"
	de "runtime/debug"
	"strings"
	"sync"
)

type namespace uint

const (
	htmlNamepsace namespace = iota
	mathmlNamespace
	svgNamespace
	xlinkNamespace
	xmlNamespace
	xmlnsNamespace
)

type quirksMode string

const (
	noQuirks      quirksMode = "no-quirks"
	quirks        quirksMode = "quirks"
	limitedQuirks quirksMode = "limited-quirks"
)

type createdByOrigin uint

const (
	htmlFragmentParsingAlgorithm createdByOrigin = iota
)

type frameset uint

const (
	framesetNotOK frameset = iota
	framesetOK
)

// HTMLTreeConstructor holds the state for various state of the tree construction phase.
type HTMLTreeConstructor struct {
	tokenChannel                                  chan *Token
	stateChannel                                  chan tokenizerState
	curInsertionMode                              insertionMode
	config                                        htmlParserConfig
	HTMLDocument                                  *spec.HTMLDocument
	quirksMode                                    quirksMode
	fosterParenting                               bool
	scriptingEnabled                              bool
	originalInsertionMode                         insertionMode
	stackOfOpenElements, activeFormattingElements spec.NodeList
	stackOfInsertionModes                         []insertionMode
	headElementPointer                            *spec.Node
	formElementPointer                            *spec.Node
	pendingTableCharacterTokens                   []*Token
	createdBy                                     createdByOrigin
	wg                                            *sync.WaitGroup
	frameset                                      frameset
	mappings                                      map[insertionMode]treeConstructionModeHandler
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

// NewHTMLTreeConstructor creates an HTMLTreeConstructor.
func NewHTMLTreeConstructor(c chan *Token, s chan tokenizerState, wg *sync.WaitGroup) *HTMLTreeConstructor {
	tr := HTMLTreeConstructor{
		tokenChannel: c,
		stateChannel: s,
		HTMLDocument: spec.NewHTMLDocumentNode(),
		wg:           wg,
	}

	tr.createMappings()
	return &tr
}

func (c *HTMLTreeConstructor) createMappings() {
	c.mappings = map[insertionMode]treeConstructionModeHandler{
		initial:            c.initialModeHandler,
		beforeHTML:         c.beforeHTMLModeHandler,
		beforeHead:         c.beforeHeadModeHandler,
		inHead:             c.inHeadModeHandler,
		inHeadNoScript:     c.inHeadNoScriptModeHandler,
		afterHead:          c.afterHeadModeHandler,
		inBody:             c.inBodyModeHandler,
		text:               c.textModeHandler,
		inTable:            c.inTableModeHandler,
		inTableText:        c.inTableTextModeHandler,
		inCaption:          c.inCaptionModeHandler,
		inColumnGroup:      c.inColumnGroupModeHandler,
		inTableBody:        c.inTableBodyModeHandler,
		inRow:              c.inRowModeHandler,
		inCell:             c.inCellModeHandler,
		inSelect:           c.inSelectModeHandler,
		inSelectInTable:    c.inSelectInTableModeHandler,
		inTemplate:         c.inTemplateModeHandler,
		afterBody:          c.afterBodyModeHandler,
		inFrameset:         c.inFramesetModeHandler,
		afterFrameset:      c.afterFramesetModeHandler,
		afterAfterBody:     c.afterAfterBodyModeHandler,
		afterAfterFrameset: c.afterAfterFramesetModeHandler,
	}
}

func (c *HTMLTreeConstructor) getCurrentNode() *spec.Node {
	if len(c.stackOfOpenElements) == 0 {
		return nil
	}
	return c.stackOfOpenElements[len(c.stackOfOpenElements)-1]
}

// Inserts a comment at a specific location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertCommentAt(t *Token, il *insertionLocation) {
	commentNode := spec.NewComment(webidl.DOMString(t.Data), il.node.OwnerDocument)
	il.insert(commentNode)
}

// Inserts a comment at the adjusted insertion location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertComment(t *Token) {
	c.insertCommentAt(t, c.getAppropriatePlaceForInsertion(nil))
}

func (c *HTMLTreeConstructor) getLastElemInStackOfOpenElements(elem webidl.DOMString) int {
	if len(c.stackOfOpenElements) == 0 {
		return -1
	}
	for i := len(c.stackOfOpenElements) - 1; i >= 0; i-- {
		if c.stackOfOpenElements[i].NodeName == elem {
			return i
		}
	}
	return -1
}

type insertionLocation struct {
	node   *spec.Node
	insert func(*spec.Node)
}

// https://html.spec.whatwg.org/multipage/parsing.html#appropriate-place-for-inserting-a-node
func (c *HTMLTreeConstructor) getAppropriatePlaceForInsertion(target *spec.Node) *insertionLocation {
	if target == nil {
		target = c.getCurrentNode()
		if target == nil {
			target = c.HTMLDocument.Node
		}
	}
	ail := &insertionLocation{}
	if c.fosterParenting {
		switch target.NodeName {
		case "table", "tbody", "tfoot", "thead", "tr":
			lastTemplate := c.getLastElemInStackOfOpenElements("template")
			lastTable := c.getLastElemInStackOfOpenElements("table")

			if lastTemplate != -1 {
				if lastTable == -1 || lastTemplate > lastTable {
					ail.node = c.stackOfOpenElements[lastTemplate]
					ail.insert = func(n *spec.Node) {
						c.stackOfOpenElements[lastTemplate].AppendChild(n)
					}
					return ail
				}
			}

			if lastTable == -1 {
				ail.node = c.stackOfOpenElements[0]
				ail.insert = func(n *spec.Node) {
					c.stackOfOpenElements[0].AppendChild(n)
				}
				return ail
			}

			if c.stackOfOpenElements[lastTable].ParentNode != nil {
				ail.node = c.stackOfOpenElements[lastTable].ParentNode
				ail.insert = func(n *spec.Node) {
					c.stackOfOpenElements[lastTable].ParentNode.InsertBefore(n, c.stackOfOpenElements[lastTable])
				}
				return ail
			}

			ail.insert = func(n *spec.Node) {
				c.stackOfOpenElements[lastTable-1].AppendChild(n)
			}
			return ail
		}
	} else {
		ail.node = target
		ail.insert = func(n *spec.Node) { target.AppendChild(n) }
	}

	return ail
}

func (c *HTMLTreeConstructor) elementInSpecificScope(target *spec.Node, list []webidl.DOMString) bool {
	for i := len(c.stackOfOpenElements) - 1; i >= 0; i-- {
		entry := c.stackOfOpenElements[i]
		if target == entry {
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

func (c *HTMLTreeConstructor) elementInScope(target *spec.Node) bool {
	list := []webidl.DOMString{
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

	return c.elementInSpecificScope(target, list)
}

func (c *HTMLTreeConstructor) elementInSelectScope(target *spec.Node) bool {
	list := []webidl.DOMString{
		"select",
	}
	return c.elementInScope(target) && c.elementInSpecificScope(target, list)
}

func (c *HTMLTreeConstructor) elementInButtonScope(target *spec.Node) bool {
	list := []webidl.DOMString{
		"button",
	}
	return c.elementInScope(target) && c.elementInSpecificScope(target, list)
}

func (c *HTMLTreeConstructor) elementInTableScope(target *spec.Node) bool {
	list := []webidl.DOMString{
		"html",
		"table",
		"template",
	}
	return c.elementInSpecificScope(target, list)
}

type validCustomElementName string

// https://html.spec.whatwg.org/multipage/custom-elements.html#custom-element-definition
type CustomElementDefinition struct {
	name               validCustomElementName
	localName          string
	observedAttributes []webidl.DOMString
	lifecycleCallbacks map[string]string
}

//https://html.spec.whatwg.org/multipage/custom-elements.html#look-up-a-custom-element-definition
func (c *HTMLTreeConstructor) lookUpCustomElementDefinition(document *spec.Node, ns, localName, is webidl.DOMString) *CustomElementDefinition {

	// browsing context
	// custom element registry
	return nil
}

func (c *HTMLTreeConstructor) clearListOfActiveFormattingElementsToLastMarker() {
	for {
		node := spec.Pop(&c.activeFormattingElements)
		if node.NodeType == spec.ScopeMarkerNode {
			return
		}
	}
}

func (c *HTMLTreeConstructor) resetInsertionMode() insertionMode {
	last := false
	lastID := len(c.stackOfOpenElements) - 1
	node := c.stackOfOpenElements[lastID]
	j := lastID
	for {
		if j == 0 {
			last = true
		}
		switch node.NodeName {
		case "select":
			if last {
				return inSelect
			}
			i := lastID
			for {
				if i == 0 {
					break
				}

				i--
				ancestor := c.stackOfOpenElements[i]
				if ancestor.NodeName == "template" {
					break
				}
				if ancestor.NodeName == "table" {
					return inSelectInTable
				}
			}

			return inSelect
		case "td", "th":
			return inCell
		case "tr":
			return inRow
		case "tbody", "thead", "tfoot":
			return inTableBody
		case "caption":
			return inCaption
		case "colgroup":
			return inColumnGroup
		case "table":
			return inTable
		case "template":
			return c.stackOfInsertionModes[len(c.stackOfInsertionModes)-1]
		case "head":
			return inHead
		case "body":
			return inBody
		case "frameset":
			return inFrameset
		case "html":
			if c.headElementPointer == nil {
				return beforeHead
			}
			return afterHead
		}

		if last {
			return inBody
		}
		j--
		node = c.stackOfOpenElements[j]
	}
}

// https://html.spec.whatwg.org/multipage/parsing.html#stop-parsing
func (c *HTMLTreeConstructor) stopParsing(err parseError) (bool, insertionMode, parseError) {
	for len(c.stackOfOpenElements) > 0 {
		spec.Pop(&c.stackOfOpenElements)
	}

	return false, stopParser, err
}

// createElementForToken creates an element from a token with the provided
// namespace and parent element.
// https://html.spec.whatwg.org/multipage/parsing.html#create-an-element-for-the-token
func (c *HTMLTreeConstructor) createElementForToken(t *Token, ns webidl.DOMString, ip *spec.Node) *spec.Node {
	document := ip.OwnerDocument
	localName := webidl.DOMString(t.TagName)
	is := t.Attributes["is"]
	// won't need to implement this for a while
	definition := c.lookUpCustomElementDefinition(document, ns, localName, webidl.DOMString(is))
	executeScript := false
	if definition != nil && c.createdBy == htmlFragmentParsingAlgorithm {
		executeScript = true
	}

	if executeScript {
		//TODO: executeScript
	}

	element := spec.NewDOMElement(document, localName, ns)
	element.Attributes = spec.NewNamedNodeMap(t.Attributes)
	element.ParentNode = ip.ParentNode
	return element
}

func (c *HTMLTreeConstructor) insertCharacter(t *Token) {
	il := c.getAppropriatePlaceForInsertion(nil)
	if il.node != nil && il.node.NodeType == spec.DocumentNode {
		return
	}

	if il.node.LastChild != nil && il.node.LastChild.NodeType == spec.TextNode {
		il.node.LastChild.Text.CharacterData.Data += webidl.DOMString(t.Data)
	} else {
		tn := spec.NewTextNode(il.node.OwnerDocument, t.Data)
		il.insert(tn)
	}

}

func (c *HTMLTreeConstructor) insertHTMLElementForToken(t *Token) *spec.Node {
	return c.insertForeignElementForToken(t, "html")
}

func (c *HTMLTreeConstructor) insertForeignElementForToken(t *Token, namespace webidl.DOMString) *spec.Node {
	il := c.getAppropriatePlaceForInsertion(nil)
	elem := c.createElementForToken(t, namespace, il.node)
	il.insert(elem)
	spec.Push(&c.stackOfOpenElements, elem)
	return elem
}

func (c *HTMLTreeConstructor) useRulesFor(t *Token, returnState, expectedState insertionMode) (bool, insertionMode, parseError) {
	reprocess, nextstate, err := c.mappings[expectedState](t)

	// if the next state is the same as the expected state, this means that mode handler didn't
	// change the state. We should use the current return state.
	if nextstate == expectedState {
		return reprocess, returnState, err
	}
	return reprocess, nextstate, err
}

func compareLastN(n int, elems spec.NodeList, elem *spec.Node) bool {
	last := len(elems) - 1
	lastElem := elems[last]
	for i := last - 1; i >= last-n; i-- {
		if elems[i].TagName != lastElem.TagName {
			return false
		}

		if elems[i].Element.NamespaceURI != lastElem.Element.NamespaceURI {
			return false
		}

		if elems[i].Attributes.Length != lastElem.Attributes.Length {
			return false
		}

		for j := 0; j < lastElem.Attributes.Length; j++ {
			if lastElem.Attributes.Item(j).NamespaceURI != elems[i].Attributes.Item(j).NamespaceURI {
				return false
			}

			if lastElem.Attributes.Item(j).Name != elems[i].Attributes.Item(j).Name {
				return false
			}

			if lastElem.Attributes.Item(j).Value != elems[i].Attributes.Item(j).Value {
				return false
			}
		}
	}

	return true
}

func (c *HTMLTreeConstructor) pushActiveFormattingElements(elem *spec.Node) {
	elems := 0

	// find the last marker
	last := len(c.activeFormattingElements) - 1
	for i := last; i >= 0 && elems < 3; i-- {
		if c.activeFormattingElements[i].NodeType == spec.ScopeMarkerNode {
			break
		}
		elems++
	}

	// elems after last marker
	if elems >= 3 && compareLastN(3, c.activeFormattingElements, c.activeFormattingElements[last]) {
		// then remove the earliest such element from the list of active formatting elements.
		c.activeFormattingElements = c.activeFormattingElements[:last]
	}

	spec.Push(&c.activeFormattingElements, elem)
}

func isSpecial(n *spec.Node) bool {
	switch n.NodeName {
	case "address", "applet", "area", "article", "aside", "base", "basefont", "bgsound", "blockquote",
		"body", "br", "button", "caption", "center", "col", "colgroup", "dd", "details", "dir", "div",
		"dl", "dt", "embed", "fieldset", "figcaption", "figure", "footer", "form", "frame", "frameset",
		"h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "iframe", "img",
		"input", "keygen", "li", "link", "listing", "main", "marquee", "menu", "meta", "nav", "noembed",
		"noframes", "noscript", "object", "ol", "p", "param", "plaintext", "pre", "script", "section",
		"select", "source", "style", "summary", "table", "tbody", "td", "template", "textarea", "tfoot",
		"th", "thead", "tr", "track", "ul", "wbr", "mi", "mo", "mn", "ms", "mtext", "annotation-xml",
		"foreignObject", "desc", "title":
		return true
	}
	return false
}

func (c *HTMLTreeConstructor) generateImpliedEndTags(blacklist []webidl.DOMString) {
	for {
		nn := c.getCurrentNode().NodeName
		switch nn {
		case "dd", "dt", "li", "optgroup", "option", "p", "rb", "rt", "rtc":
			for _, b := range blacklist {
				if b == nn {
					return
				}
			}
			spec.Pop(&c.stackOfOpenElements)
			continue
		}
		break
	}
}

func (c *HTMLTreeConstructor) closePElement() {
	c.generateImpliedEndTags([]webidl.DOMString{"p"})
	nn := c.getCurrentNode().NodeName
	if nn != "p" {
		// parser error
	}
	for {
		p := spec.Pop(&c.stackOfOpenElements)
		if p.NodeName == "p" {
			break
		}
	}
}

func (c *HTMLTreeConstructor) adoptionAgencyAlgorithm(t *Token) (bool, parseError) {
	var err parseError
	cur := c.getCurrentNode()
	if cur.NodeName == webidl.DOMString(t.TagName) && spec.Contains(cur, &c.activeFormattingElements) == -1 {
		spec.Pop(&c.stackOfOpenElements)
		return false, noError
	}

	// outer loop
	var y, z, si, nif, nis int
	for x := 1; x < 8; x++ {
		// 6
		var formattingElement *spec.Node
		for y = len(c.activeFormattingElements) - 1; y >= 0; y-- {
			if c.activeFormattingElements[y].NodeType == spec.ScopeMarkerNode {
				break
			}

			if c.activeFormattingElements[y].NodeName == webidl.DOMString(t.TagName) {
				formattingElement = c.activeFormattingElements[y]
				break
			}
		}

		if formattingElement == nil {
			return true, noError
		}

		// 7
		si = spec.Contains(formattingElement, &c.stackOfOpenElements)
		if si == -1 {
			// parse error
			spec.Remove(y, &c.activeFormattingElements)
			return false, noError
		}

		// 8
		if !c.elementInScope(formattingElement) {
			// parse error
			return false, generalParseError
		}

		// 9
		if formattingElement != c.getCurrentNode() {
			err = generalParseError
		}

		// 10
		var furthestBlock *spec.Node
		for z = si + 1; z < len(c.stackOfOpenElements); z++ {
			if isSpecial(c.stackOfOpenElements[z]) {
				furthestBlock = c.stackOfOpenElements[z]
				break
			}
		}

		// 11
		if furthestBlock == nil {
			for {
				if c.getCurrentNode() == formattingElement {
					spec.Pop(&c.stackOfOpenElements)
					spec.Remove(y, &c.activeFormattingElements)
					return false, noError
				}
				spec.Pop(&c.stackOfOpenElements)
			}
		}

		// 12
		ca := c.stackOfOpenElements[si-1]
		// 13
		bm := c.activeFormattingElements[y]

		//14 inner loop
		node := furthestBlock
		lastNode := furthestBlock
		for a := 1; z >= 0; a++ {
			z--
			// 14.3
			node = c.stackOfOpenElements[z]

			// 14.4
			if node == formattingElement {
				break
			}

			// 14.5
			if a > 3 {
				spec.Remove(spec.Contains(node, &c.activeFormattingElements), &c.activeFormattingElements)
			}

			// 14.6
			if spec.Contains(node, &c.activeFormattingElements) != -1 {
				spec.Remove(spec.Contains(node, &c.stackOfOpenElements), &c.stackOfOpenElements)
				continue
			}

			// 14.7
			elem := c.createElementForToken(t, "html", ca)
			nif = spec.Contains(node, &c.activeFormattingElements)
			if nif != -1 {
				c.activeFormattingElements[nif] = elem
			}
			nis = spec.Contains(node, &c.stackOfOpenElements)
			if nis != -1 {
				c.stackOfOpenElements[nis] = elem
			}

			// 14.8
			if lastNode == furthestBlock {
				bm = c.activeFormattingElements[nif+1]
			}
			// 14.9
			lastNode.ParentNode.RemoveChild(lastNode)
			node.AppendChild(lastNode)
			// 14.10
			lastNode = node
		}

		// 15
		// this step is NOT explicitly stated but I think it is implied otherwise, this algorithm
		// doesn't really do anything.
		if lastNode.ParentNode != nil {
			lastNode.ParentNode.RemoveChild(lastNode)
		}
		ca.AppendChild(lastNode)

		// 16
		clone := formattingElement.CloneNode(true)
		clone.ParentNode = furthestBlock
		// 17
		for i, child := range furthestBlock.ChildNodes {

			// same as above, here. This step does NOT explicitly say to remove the children elements
			// and move them, but the algorithm wouldn't really work otherwise.
			spec.Remove(i, &furthestBlock.ChildNodes)
			clone.AppendChild(child)
		}
		// 18
		furthestBlock.AppendChild(clone)
		// 19
		f := spec.Contains(formattingElement, &c.activeFormattingElements)
		if f != -1 {
			spec.Remove(f, &c.activeFormattingElements)
			b := spec.Contains(bm, &c.activeFormattingElements)
			if b != -1 {
				c.activeFormattingElements[b] = clone
			}
		}

		//20
		f = spec.Contains(formattingElement, &c.stackOfOpenElements)
		if f != -1 {
			spec.Remove(f, &c.stackOfOpenElements)
			b := spec.Contains(furthestBlock, &c.stackOfOpenElements)
			if b != -1 {
				if b+1 == len(c.stackOfOpenElements) {
					spec.Push(&c.stackOfOpenElements, clone)
				} else {
					c.stackOfOpenElements[b+1] = clone
				}
			}
		}
	}

	return false, err
}

func (c *HTMLTreeConstructor) racfeCreateStep(i int) {
	// 8. Create: Insert an HTML element for the token for which the element entry was created,
	// to obtain new element.
	il := c.getAppropriatePlaceForInsertion(nil)
	elem := c.activeFormattingElements[i].CloneNode(true)
	il.insert(elem)
	spec.Push(&c.stackOfOpenElements, elem)

	// 9. Replace the entry for entry in the list with an entry for new element.
	c.activeFormattingElements[i] = elem

	// 10. If the entry for new element in the list of active formatting elements is not the last
	// entry in the list, return to the step labeled advance.
}

func (c *HTMLTreeConstructor) reconstructActiveFormattingElements() {
	// 1. If there are no entries in the list of active formatting elements, then there is nothing
	// to reconstruct; stop this algorithm.
	if len(c.activeFormattingElements) == 0 {
		return
	}

	// 2. If the last (most recently added) entry in the list of active formatting elements is a
	// marker, or if it is an element that is in the stack of open elements, then there is nothing
	// to reconstruct; stop this algorithm.
	last := len(c.activeFormattingElements) - 1
	lafe := c.activeFormattingElements[last]
	doesContain := spec.Contains(lafe, &c.stackOfOpenElements)
	if lafe.NodeType == spec.ScopeMarkerNode || doesContain != -1 {
		return
	}

	// 3. Let entry be the last (most recently added) element in the list of active formatting
	// elements.
	i := last

	// 4. Rewind: If there are no entries before entry in the list of active formatting elements,
	// then jump to the step labeled create.
	if i == 0 {
		c.racfeCreateStep(i)
	} else {
		// 5. Let entry be the entry one earlier than entry in the list of active formatting elements.
		for ; i > 0; i-- {
			// 6. If entry is neither a marker nor an element that is also in the stack of open elements,
			// go to the step labeled rewind.
			doesContain = spec.Contains(c.activeFormattingElements[i], &c.stackOfOpenElements)
			if c.activeFormattingElements[i].NodeType == spec.ScopeMarkerNode || doesContain != -1 {
				break
			}
		}
	}

	// 7. Advance: Let entry be the element one later than entry in the list of active formatting
	// elements.
	for j := i + 1; j < len(c.activeFormattingElements); j++ {
		c.racfeCreateStep(j)
	}

}

func (c *HTMLTreeConstructor) clearStackBackToTable() {
	for {
		switch c.getCurrentNode().NodeName {
		case "table", "template", "html":
			return
		}
		spec.Pop(&c.stackOfOpenElements)
	}
}

func (c *HTMLTreeConstructor) clearStackBackToTableRow() {
	for {
		switch c.getCurrentNode().NodeName {
		case "tr", "template", "html":
			return
		}
		spec.Pop(&c.stackOfOpenElements)
	}
}

func (c *HTMLTreeConstructor) clearStackBackToTableBody() {
	for {
		switch c.getCurrentNode().NodeName {
		case "tbody", "tfoot", "thead", "template", "html":
			return
		}
		spec.Pop(&c.stackOfOpenElements)
	}
}

const w30DTDW3HTMLStrict3En string = "-//W3O//DTD W3 HTML Strict 3.0//EN//"
const w3cDTDHTML4TransitionalEN string = "-/W3C/DTD HTML 4.0 Transitional/EN"
const htmlString string = "HTML"
const ibmxhtml string = "http://www.ibm.com/data/dtd/v11/ibmxhtml1-transitional.dtd"

const silmarilDTDHTMLPro string = "+//Silmaril//dtd html Pro v0r11 19970101//"
const dTDHTML3asWedit string = "-//AS//DTD HTML 3.0 asWedit + extensions//"
const advaSoftDTDHTML3 string = "-//AdvaSoft Ltd//DTD HTML 3.0 asWedit + extensions//"
const iETFDTDHTML2Level1 string = "-//IETF//DTD HTML 2.0 Level 1//"
const iETFDTDHTML2Level2 string = "-//IETF//DTD HTML 2.0 Level 2//"
const iETFDTDHTML2StrictLevel1 string = "-//IETF//DTD HTML 2.0 Strict Level 1//"
const iETFDTDHTML2StrictLevel2 string = "-//IETF//DTD HTML 2.0 Strict Level 2//"
const iETFDTDHTML2Strict string = "-//IETF//DTD HTML 2.0 Strict//"
const iETFDTDHTML2 string = "-//IETF//DTD HTML 2.0//"
const iIETFDTDHTML2E string = "-//IETF//DTD HTML 2.1E//"
const iETFDTDHTML30 string = "-//IETF//DTD HTML 3.0//"
const iETFDTDHTML32Final string = "-//IETF//DTD HTML 3.2 Final//"
const iETFDTDHTML32 string = "-//IETF//DTD HTML 3.2//"
const iETFDTDHTML3 string = "-//IETF//DTD HTML 3//"
const iETFDTDHTMLLevel0 string = "-//IETF//DTD HTML Level 0//"
const iETFDTDHTMLLevel1 string = "-//IETF//DTD HTML Level 1//"
const iETFDTDHTMLLevel2 string = "-//IETF//DTD HTML Level 2//"
const iETFDTDHTMLLevel3 string = "-//IETF//DTD HTML Level 3//"
const iETFDTDHTMLStrictLevel0 string = "-//IETF//DTD HTML Strict Level 0//"
const iETFDTDHTMLStrictLevel1 string = "-//IETF//DTD HTML Strict Level 1//"
const iETFDTDHTMLStrictLevel2 string = "-//IETF//DTD HTML Strict Level 2//"
const iETFDTDHTMLStrictLevel3 string = "-//IETF//DTD HTML Strict Level 3//"
const iETFDTDHTMLStrict string = "-//IETF//DTD HTML Strict//"
const iETFDTDHTML string = "-//IETF//DTD HTML//"
const metriusDTDMetriusPresentational string = "-//Metrius//DTD Metrius Presentational//"
const microsoftDTDInternetExplorer2HTMLStrict string = "-//Microsoft//DTD Internet Explorer 2.0 HTML Strict//"
const microsoftDTDInternetExplorer2HTML string = "-//Microsoft//DTD Internet Explorer 2.0 HTML//"
const microsoftDTDInternetExplorer2Tables string = "-//Microsoft//DTD Internet Explorer 2.0 Tables//"
const microsoftDTDInternetExplorer3HTMLStrict string = "-//Microsoft//DTD Internet Explorer 3.0 HTML Strict//"
const microsoftDTDInternetExplorer3HTML string = "-//Microsoft//DTD Internet Explorer 3.0 HTML//"
const microsoftDTDInternetExplorer3Tables string = "-//Microsoft//DTD Internet Explorer 3.0 Tables//"
const netscapeCommCorpDTDHTML string = "-//Netscape Comm. Corp.//DTD HTML//"
const netscapeCommCorpDTDStrictHTML string = "-//Netscape Comm. Corp.//DTD Strict HTML//"
const oReillyAssociatesDTDHTML2 string = "-//O'Reilly and Associates//DTD HTML 2.0//"
const oReillyAssociatesDTDHTMLExtended1 string = "-//O'Reilly and Associates//DTD HTML Extended 1.0//"
const oReillyAssociatesDTDHTMLExtendedRelaxed1 string = "-//O'Reilly and Associates//DTD HTML Extended Relaxed 1.0//"
const sQDTDHTML2HoTMetaLExtensions string = "-//SQ//DTD HTML 2.0 HoTMetaL + extensions//"
const softQuadSoftwareDTDHoTMetaLPRO string = "-//SoftQuad Software//DTD HoTMetaL PRO 6.0::19990601::extensions to HTML 4.0//"
const softQuadDTDHoTMetaLPRO string = "-//SoftQuad//DTD HoTMetaL PRO 4.0::19971010::extensions to HTML 4.0//"
const spyglassDTDHTML2Extended string = "-//Spyglass//DTD HTML 2.0 Extended//"
const sunMicrosystemsCorpDTDHotJavaHTML string = "-//Sun Microsystems Corp.//DTD HotJava HTML//"
const sunMicrosystemsCorpDTDHotJavaStrictHTML string = "-//Sun Microsystems Corp.//DTD HotJava Strict HTML//"
const w3cDTDHTML31 string = "-//W3C//DTD HTML 3 1995-03-24//"
const w3cDTDHTML32Draft string = "-//W3C//DTD HTML 3.2 Draft//"
const w3cDTDHTML32Final string = "-//W3C//DTD HTML 3.2 Final//"
const w3cDTDHTML32 string = "-//W3C//DTD HTML 3.2//"
const w3cDTDHTML32SDraft string = "-//W3C//DTD HTML 3.2S Draft//"
const w3cDTDHTML4Frameset string = "-//W3C//DTD HTML 4.0 Frameset//"
const w3cDTDHTML4Transitional string = "-//W3C//DTD HTML 4.0 Transitional//"
const w3cDTDHTML401Frameset string = "-//W3C//DTD HTML 4.01 Frameset//"
const w3cDTDHTML401Transitional string = "-//W3C//DTD HTML 4.01 Transitional//"
const w3cDTDHTMLExperimental1996 string = "-//W3C//DTD HTML Experimental 19960712//"
const w3cDTDHTMLExperimental9704 string = "-//W3C//DTD HTML Experimental 970421//"
const w3cDTDXHTML1Frameset string = "-//W3C//DTD XHTML 1.0 Frameset//"
const w3cDTDXHTML1Transitional string = "-//W3C//DTD XHTML 1.0 Transitional//"
const w3cDTDW3HTML string = "-//W3C//DTD W3 HTML//"
const w3cDTDW3HTML3 string = "-//W3O//DTD W3 HTML 3.0//"
const webTechsDTDMozillaHTML2 string = "-//WebTechs//DTD Mozilla HTML 2.0//"
const webTechsDTDMozillaHTML string = "-//WebTechs//DTD Mozilla HTML//"

var knownPublicIdentifiers = []string{
	silmarilDTDHTMLPro,
	dTDHTML3asWedit,
	advaSoftDTDHTML3,
	iETFDTDHTML2Level1,
	iETFDTDHTML2Level2,
	iETFDTDHTML2StrictLevel1,
	iETFDTDHTML2StrictLevel2,
	iETFDTDHTML2Strict,
	iETFDTDHTML2,
	iIETFDTDHTML2E,
	iETFDTDHTML30,
	iETFDTDHTML32Final,
	iETFDTDHTML32,
	iETFDTDHTML3,
	iETFDTDHTMLLevel0,
	iETFDTDHTMLLevel1,
	iETFDTDHTMLLevel2,
	iETFDTDHTMLLevel3,
	iETFDTDHTMLStrictLevel0,
	iETFDTDHTMLStrictLevel1,
	iETFDTDHTMLStrictLevel2,
	iETFDTDHTMLStrictLevel3,
	iETFDTDHTMLStrict,
	iETFDTDHTML,
	metriusDTDMetriusPresentational,
	microsoftDTDInternetExplorer2HTMLStrict,
	microsoftDTDInternetExplorer2HTML,
	microsoftDTDInternetExplorer2Tables,
	microsoftDTDInternetExplorer3HTMLStrict,
	microsoftDTDInternetExplorer3HTML,
	microsoftDTDInternetExplorer3Tables,
	netscapeCommCorpDTDHTML,
	netscapeCommCorpDTDStrictHTML,
	oReillyAssociatesDTDHTML2,
	oReillyAssociatesDTDHTMLExtended1,
	oReillyAssociatesDTDHTMLExtendedRelaxed1,
	sQDTDHTML2HoTMetaLExtensions,
	softQuadSoftwareDTDHoTMetaLPRO,
	softQuadDTDHoTMetaLPRO,
	spyglassDTDHTML2Extended,
	sunMicrosystemsCorpDTDHotJavaHTML,
	sunMicrosystemsCorpDTDHotJavaStrictHTML,
	w3cDTDHTML31,
	w3cDTDHTML32Draft,
	w3cDTDHTML32Final,
	w3cDTDHTML32,
	w3cDTDHTML32SDraft,
	w3cDTDHTML4Frameset,
	w3cDTDHTML4Transitional,
	w3cDTDHTMLExperimental1996,
	w3cDTDHTMLExperimental9704,
	w3cDTDW3HTML,
	w3cDTDW3HTML3,
	webTechsDTDMozillaHTML2,
	webTechsDTDMozillaHTML,
}

func (c *HTMLTreeConstructor) isIframeSrcDoc() bool {
	return false
}

func (c *HTMLTreeConstructor) isForceQuirks(t *Token) bool {
	if !c.isIframeSrcDoc() {
		if t.ForceQuirks {
			return true
		}

		if t.TagName != "html" {
			return true
		}
		switch t.PublicIdentifier {
		case w30DTDW3HTMLStrict3En, w3cDTDHTML4TransitionalEN, htmlString:
			return true
		default:
			if t.SystemIdentifier == ibmxhtml {
				return true
			}

			for _, v := range knownPublicIdentifiers {
				if strings.HasPrefix(t.PublicIdentifier, v) {
					return true
				}
			}

			if (t.SystemIdentifier == missing &&
				strings.HasPrefix(t.PublicIdentifier, w3cDTDHTML401Frameset)) ||
				(t.SystemIdentifier == missing && strings.HasPrefix(t.PublicIdentifier, w3cDTDHTML401Transitional)) {
				return true
			}
		}
	}
	return false
}

func (c *HTMLTreeConstructor) isLimitedQuirks(t *Token) bool {
	if strings.HasPrefix(t.PublicIdentifier, w3cDTDXHTML1Frameset) {
		return true
	}

	if strings.HasPrefix(t.PublicIdentifier, w3cDTDXHTML1Transitional) {
		return true
	}

	if t.SystemIdentifier != missing {

		if strings.HasPrefix(t.PublicIdentifier, w3cDTDHTML401Frameset) {
			return true
		}
		if strings.HasPrefix(t.PublicIdentifier, w3cDTDHTML401Transitional) {
			return true
		}
	}
	return false
}

func (c *HTMLTreeConstructor) defaultInitialModeHandler() (bool, insertionMode, parseError) {
	return true, beforeHTML, noError
}

// https://html.spec.whatwg.org/multipage/parsing.html#the-initial-insertion-mode
func (c *HTMLTreeConstructor) initialModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, initial, noError
		}
	case commentToken:
		c.insertComment(t)
		return false, initial, noError
	case docTypeToken:
		if t.TagName != "html" ||
			t.PublicIdentifier != missing ||
			(t.SystemIdentifier != missing &&
				t.SystemIdentifier != "about:legacy-compat") {
			err = generalParseError
		}

		doctype := spec.NewDocTypeNode(t.TagName, t.PublicIdentifier, t.SystemIdentifier)
		c.HTMLDocument.AppendChild(doctype)
		c.HTMLDocument.Node.Document.Doctype = doctype

		if c.isForceQuirks(t) {
			c.quirksMode = quirks
		} else if c.isLimitedQuirks(t) {
			c.quirksMode = limitedQuirks
		} else {
			c.quirksMode = noQuirks
		}

		return false, beforeHTML, err
	}
	return c.defaultInitialModeHandler()
}

func (c *HTMLTreeConstructor) defaultBeforeHTMLModeHandler(t *Token) (bool, insertionMode, parseError) {
	n := spec.NewDOMElement(c.HTMLDocument.Node, "html", "html")
	n.OwnerDocument = c.HTMLDocument.Node
	c.HTMLDocument.AppendChild(n)
	spec.Push(&c.stackOfOpenElements, n)

	// TODO: application cache algo
	return true, beforeHead, noError
}

// https://html.spec.whatwg.org/multipage/parsing.html#the-before-html-insertion-mode
func (c *HTMLTreeConstructor) beforeHTMLModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case docTypeToken:
		return false, beforeHTML, generalParseError
	case commentToken:
		//c.insertCommentAt(t, lastChildOfDocument)
		return false, beforeHTML, noError
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, beforeHTML, noError
		}
	case startTagToken:
		if t.TagName == "html" {
			elem := c.createElementForToken(t, "html", c.HTMLDocument.Node)
			c.HTMLDocument.AppendChild(elem)
			spec.Push(&c.stackOfOpenElements, elem)
			// handle navigation of a browsing context
			return false, beforeHead, noError
		}
	case endTagToken:
		switch t.TagName {
		case "head", "body", "html", "br":
			return c.defaultBeforeHTMLModeHandler(t)
		default:
			return false, beforeHTML, generalParseError
		}
	}
	return c.defaultBeforeHTMLModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultBeforeHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	tok := &Token{
		TokenType: startTagToken,
		TagName:   "head",
	}
	elem := c.insertHTMLElementForToken(tok)
	c.headElementPointer = elem
	return true, inHead, noError
}
func (c *HTMLTreeConstructor) beforeHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, beforeHead, noError
		}
	case commentToken:
		c.insertComment(t)
		return false, beforeHead, noError
	case docTypeToken:
		return false, beforeHead, generalParseError
	case startTagToken:
		if t.TagName == "html" {
			return c.useRulesFor(t, beforeHead, inBody)
		}

		if t.TagName == "head" {
			elem := c.insertHTMLElementForToken(t)
			c.headElementPointer = elem
			return false, inHead, noError
		}
	case endTagToken:
		switch t.TagName {
		case "head", "body", "html", "br":
			return c.defaultBeforeHeadModeHandler(t)
		}

		return false, beforeHead, generalParseError
	}

	return c.defaultBeforeHeadModeHandler(t)

}

func (c *HTMLTreeConstructor) genericRCDATAElementParsingAlgorithm(t *Token, ogi insertionMode) (bool, insertionMode, parseError) {
	c.insertHTMLElementForToken(t)
	c.originalInsertionMode = ogi
	c.stateChannel <- rcDataState
	return false, text, noError
}

func (c *HTMLTreeConstructor) genericRawTextElementParsingAlgorithm(t *Token, ogi insertionMode) (bool, insertionMode, parseError) {
	c.insertHTMLElementForToken(t)
	c.originalInsertionMode = ogi
	c.stateChannel <- rawTextState
	return false, text, noError
}

func (c *HTMLTreeConstructor) defaultInHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	spec.Pop(&c.stackOfOpenElements)
	return true, afterHead, noError
}
func (c *HTMLTreeConstructor) inHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, inHead, noError
		}
	case commentToken:
		c.insertComment(t)
		return false, inHead, noError
	case docTypeToken:
		return false, inHead, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inHead, inBody)
		case "base", "basefont", "bgsound", "link":
			/*c.WriteHTMLElement(t)
			c.PopOpenElements()*/
			//TODO: acknowledge the self closing flag?
		case "meta":
			/*c.WriteHTMLElement(t)
			c.PopOpenElements()*/
			//TODO: acknowledge the self closing flag?
		case "title":
			return c.genericRCDATAElementParsingAlgorithm(t, inHead)
		case "noscript":
			if c.scriptingEnabled {

			} else {
				//c.WriteHTMLElement(t)
				return false, inHeadNoScript, noError
			}
		case "noframes", "style":
		case "script":
			il := c.getAppropriatePlaceForInsertion(nil)
			elem := c.createElementForToken(t, "html", il.node)
			elem.ParserDocument = c.HTMLDocument
			elem.NonBlocking = false
			il.insert(elem)
			spec.Push(&c.stackOfOpenElements, elem)
			c.stateChannel <- scriptDataState
			c.originalInsertionMode = inHead
			return false, text, noError
		case "template":
		case "head":
			return false, inHead, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "head":
			spec.Pop(&c.stackOfOpenElements)
			return false, afterHead, noError
		case "body", "html", "br":
			return c.defaultInHeadModeHandler(t)
		case "template":
		default:
			return false, inHead, generalParseError
		}
	}

	return c.defaultInHeadModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultInHeadNoScriptModeHandler(t *Token) (bool, insertionMode, parseError) {
	//c.PopOpenElements()
	return true, inHead, generalParseError
}
func (c *HTMLTreeConstructor) inHeadNoScriptModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return c.useRulesFor(t, inHeadNoScript, inHead)
		}
	case commentToken:
		return c.useRulesFor(t, inHeadNoScript, inHead)
	case docTypeToken:
		return false, inHeadNoScript, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inHeadNoScript, inBody)
		case "basefont", "bgsound", "link", "meta", "noframe", "style":
			return c.useRulesFor(t, inHeadNoScript, inHead)
		case "head", "noscript":
			return false, inHeadNoScript, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "noscript":
			//c.PopOpenElements()
			return false, inHead, noError
		case "br":
			return c.defaultInHeadNoScriptModeHandler(t)
		default:
			return false, inHeadNoScript, generalParseError
		}
	}
	return c.defaultInHeadNoScriptModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultAfterHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	tok := &Token{
		TokenType: startTagToken,
		TagName:   "body",
	}
	c.insertHTMLElementForToken(tok)
	return true, inBody, noError
}
func (c *HTMLTreeConstructor) afterHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.insertCharacter(t)
			return false, afterHead, noError
		}
	case commentToken:
		c.insertComment(t)
		return false, afterHead, noError
	case docTypeToken:
		return false, afterHead, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, afterHead, inBody)
		case "body":
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inBody, noError
		case "frameset":
			//c.WriteHTMLElement(t)
			return false, inFrameset, noError
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style", "template", "title":
			//c.PushOpenElements(c.headPointer)
			reprocess, nextmode, err := c.inHeadModeHandler(t)

			//c.PopOpenElements()
			return reprocess, nextmode, err
		case "head":
			return false, afterHead, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "template":
			return c.useRulesFor(t, afterHead, inHead)
		case "body", "html", "br":
			return c.defaultAfterHeadModeHandler(t)
		default:
			return false, afterHead, generalParseError
		}

	}
	return c.defaultAfterHeadModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultInBodyModeHandler(t *Token) parseError {
	err := noError
	for i := len(c.stackOfOpenElements) - 1; i >= 0; i-- {
		node := c.stackOfOpenElements[i]
		if node.NodeName == webidl.DOMString(t.TagName) {
			c.generateImpliedEndTags([]webidl.DOMString{node.NodeName})
			if node != c.getCurrentNode() {
				err = generalParseError
			}

			for {
				popped := spec.Pop(&c.stackOfOpenElements)
				if popped == nil || popped == node {
					return err
				}
			}
		} else {
			if isSpecial(node) {
				return generalParseError
			}
		}
	}
	return err
}

func containedIn(s webidl.DOMString, h []webidl.DOMString) bool {
	for _, t := range h {
		if s == t {
			return true
		}
	}

	return false
}

func (c *HTMLTreeConstructor) stackOfOpenElementsParseErrorCheck() parseError {
	ls := []webidl.DOMString{
		"dd",
		"dt",
		"li",
		"optgroup",
		"option",
		"p",
		"rb",
		"rp",
		"rt",
		"rtc",
		"tbody",
		"td",
		"tfoot",
		"th",
		"thead",
		"tr",
		"body",
		"html",
	}
	for _, s := range c.stackOfOpenElements {
		if !containedIn(s.TagName, ls) {
			return generalParseError
		}
	}

	return noError
}

func (c *HTMLTreeConstructor) containedInStackOpenElements(s string) []*spec.Node {
	nodes := make([]*spec.Node, 0)
	for _, o := range c.stackOfOpenElements {
		if string(o.NodeName) == s {
			nodes = append(nodes, o)
		}
	}

	return nodes
}

func (c *HTMLTreeConstructor) stackContainsInScope(s string, scopeFunc func(*spec.Node) bool, handlerFunc func(), trueOrFalseFunc bool) {
	nodes := c.containedInStackOpenElements(s)
	if len(nodes) != 0 {
		for _, n := range nodes {
			if scopeFunc(n) {
				if trueOrFalseFunc {
					handlerFunc()
				}
				return
			}
		}
	}

	if !trueOrFalseFunc {
		handlerFunc()
	}
}

func (c *HTMLTreeConstructor) inBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0000":
			return false, inBody, generalParseError
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.reconstructActiveFormattingElements()
			c.insertCharacter(t)
		default:
			c.reconstructActiveFormattingElements()
			c.insertCharacter(t)
			c.frameset = framesetNotOK
		}
	case commentToken:
		c.insertComment(t)
		return false, inBody, noError
	case docTypeToken:
		return false, inBody, generalParseError
	case startTagToken:
		switch t.TagName {
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style",
			"template", "title":
			return c.useRulesFor(t, inBody, inHead)
		case "body":
		case "frameset":
		case "address", "article", "aside", "blockquote", "center", "details", "dialog", "dir",
			"div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "main",
			"menu", "nav", "ol", "p", "section", "summary", "ul":
			c.stackContainsInScope("p", c.elementInButtonScope, c.closePElement, true)
			c.insertHTMLElementForToken(t)
		case "h1", "h2", "h3", "h4", "h5", "h6":
			err = noError
			c.stackContainsInScope("p", c.elementInButtonScope, c.closePElement, true)
			switch c.getCurrentNode().NodeName {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				err = generalParseError
				spec.Pop(&c.stackOfOpenElements)
			}

			c.insertHTMLElementForToken(t)
			return false, inBody, err
		case "pre", "listing":
		case "form":
		case "li":
		case "dd", "dt":
		case "plaintext":
			c.stackContainsInScope("button", c.elementInButtonScope, c.closePElement, true)
			c.insertHTMLElementForToken(t)
			c.stateChannel <- plaintextState
		case "button":
			c.stackContainsInScope("button", c.elementInScope, func() {
				c.generateImpliedEndTags([]webidl.DOMString{})
				for {
					popped := spec.Pop(&c.stackOfOpenElements)
					if popped == nil || popped.NodeName == "button" {
						break
					}
				}
			}, true)
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
		case "a":
			var node *spec.Node
			for i := len(c.activeFormattingElements) - 1; i >= 0; i++ {
				node = c.activeFormattingElements[i]
				if node.NodeType == spec.ScopeMarkerNode {
					break
				}

				if node.NodeName == "a" {
					var shouldDefault bool
					shouldDefault, err = c.adoptionAgencyAlgorithm(t)
					if shouldDefault {
						err = c.defaultInBodyModeHandler(t)
					}
					spec.Remove(spec.Contains(node, &c.activeFormattingElements), &c.activeFormattingElements)
					spec.Remove(spec.Contains(node, &c.stackOfOpenElements), &c.stackOfOpenElements)
					break
				}
			}

			c.reconstructActiveFormattingElements()
			elem := c.insertHTMLElementForToken(t)
			c.pushActiveFormattingElements(elem)
		case "b", "big", "code", "em", "font", "i", "s", "small", "strike", "strong", "tt", "u":
			c.reconstructActiveFormattingElements()
			elem := c.insertHTMLElementForToken(t)
			spec.Push(&c.activeFormattingElements, elem)
		case "nobr":
		case "applet", "marquee", "object":
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			spec.Push(&c.activeFormattingElements, spec.ScopeMarker)
			c.frameset = framesetNotOK
		case "table":
			if c.quirksMode != quirks {
				c.stackContainsInScope("p", c.elementInButtonScope, c.closePElement, true)
			}
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inTable, noError
		case "area", "br", "embed", "img", "keygen", "wbr":
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			spec.Pop(&c.stackOfOpenElements)
			//ack token?
			c.frameset = framesetNotOK
		case "input":
		case "param", "source", "track":
		case "hr":
			c.stackContainsInScope("p", c.elementInButtonScope, c.closePElement, true)
			c.insertHTMLElementForToken(t)
			spec.Pop(&c.stackOfOpenElements)
			// ack self closing flag
			c.frameset = framesetNotOK
		case "image":
		case "textarea":
		case "xmp":
		case "iframe":
		case "noembed":
		case "noscript":
		case "select":
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			switch c.curInsertionMode {
			case inTable, inCaption, inTableBody, inRow, inCell:
				return false, inSelectInTable, noError
			}
			return false, inSelect, noError
		case "optgroup", "option":
			if c.getCurrentNode().NodeName == "option" {
				spec.Pop(&c.stackOfOpenElements)
			}
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
		case "rb", "rtc":
		case "rp", "rt":
		case "math":
		case "svg":
		case "caption", "col", "colgroup", "frame", "head", "tbody", "td", "tfoot", "th", "thead", "tr":
		default:
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
		}
		return false, inBody, err
	case endTagToken:
		switch t.TagName {
		case "template":
		case "body":
			target := &spec.Node{
				NodeType: spec.ElementNode,
				NodeName: "body",
			}
			if !c.elementInScope(target) {
				return false, inBody, generalParseError
			}

			e := c.stackOfOpenElementsParseErrorCheck()

			return false, afterBody, e
		case "html":
		case "address", "article", "aside", "blockquote", "button", "center", "details", "dialog",
			"dir", "div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup",
			"listing", "main", "menu", "nav", "ol", "pre", "section", "summary", "ul":

			tagInScope := false
			c.stackContainsInScope(t.TagName, c.elementInScope, func() {
				tagInScope = true
			}, true)

			if !tagInScope {
				return false, inBody, generalParseError
			}

			c.generateImpliedEndTags([]webidl.DOMString{})
			if c.getCurrentNode().NodeName != webidl.DOMString(t.TagName) {
				err = generalParseError
			}
			var popped *spec.Node
			for {
				popped = spec.Pop(&c.stackOfOpenElements)
				if popped == nil || popped.NodeName == webidl.DOMString(t.TagName) {
					break
				}
			}
		case "form":
		case "p":
			tagInScope := false
			c.stackContainsInScope("p", c.elementInButtonScope, func() {
				tagInScope = true
			}, true)

			if !tagInScope {
				err = generalParseError
				tok := &Token{
					TagName:   "p",
					TokenType: startTagToken,
				}
				c.insertHTMLElementForToken(tok)
			}
			c.closePElement()
		case "li":
		case "dd", "dt":
		case "h1", "h2", "h3", "h4", "h5", "h6":
		case "sarcasm":
		case "a", "b", "big", "code", "em", "font", "i", "nobr", "s", "small", "strike", "strong",
			"tt", "u":
			var shouldDefault bool
			shouldDefault, err = c.adoptionAgencyAlgorithm(t)
			if shouldDefault {
				err = c.defaultInBodyModeHandler(t)
			}
		case "applet", "marquee", "object":
		case "br":
		default:
			err = c.defaultInBodyModeHandler(t)
		}

		return false, inBody, err
	case endOfFileToken:
		if len(c.stackOfOpenElements) != 0 {
			return c.useRulesFor(t, inBody, inTemplate)
		}

		err := c.stackOfOpenElementsParseErrorCheck()
		return c.stopParsing(err)
	}
	return false, inBody, noError
}
func (c *HTMLTreeConstructor) textModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		c.insertCharacter(t)
		return false, text, noError
	case endOfFileToken:
		node := c.getCurrentNode()
		if node.NodeName == "script" {
			node.AlreadyStated = true
		}
		spec.Pop(&c.stackOfOpenElements)
		return true, c.originalInsertionMode, generalParseError
	case endTagToken:
		switch t.TagName {
		case "script":
			spec.Pop(&c.stackOfOpenElements)
			return false, c.originalInsertionMode, noError
		default:
			spec.Pop(&c.stackOfOpenElements)
			return false, c.originalInsertionMode, noError
		}
	}
	return false, text, noError
}

func (c *HTMLTreeConstructor) defaultInTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	c.fosterParenting = true
	repro, nextState, _ := c.useRulesFor(t, inTable, inBody)
	c.fosterParenting = false
	return repro, nextState, generalParseError
}
func (c *HTMLTreeConstructor) inTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch c.getCurrentNode().NodeName {
		case "table", "tbody", "tfoot", "thead", "tr":
			c.pendingTableCharacterTokens = []*Token{}
			c.originalInsertionMode = c.curInsertionMode
			return true, inTableText, noError
		}
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		return false, inTable, generalParseError
	case startTagToken:
		switch t.TagName {
		case "caption":
			c.clearStackBackToTable()
			spec.Push(&c.activeFormattingElements, spec.ScopeMarker)
			c.insertHTMLElementForToken(t)
			return false, inCaption, noError
		case "colgroup":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(t)
			return false, inColumnGroup, noError
		case "col":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(&Token{
				TagName:   "colgroup",
				TokenType: startTagToken,
			})
			return true, inColumnGroup, noError
		case "tbody", "tfoot", "thead":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(t)
			return false, inTableBody, noError
		case "td", "th", "tr":
			c.clearStackBackToTable()
			tbody := &Token{
				TokenType: startTagToken,
				TagName:   "tbody",
			}
			c.insertHTMLElementForToken(tbody)
			return true, inTableBody, noError
		case "table":
			repro := false
			mode := inTable
			c.stackContainsInScope("table", c.elementInTableScope, func() {
				spec.PopUntil(&c.stackOfOpenElements, "table")
				mode = c.resetInsertionMode()
				repro = true
			}, true)
			return repro, mode, noError
		case "style", "script", "template":
			return c.useRulesFor(t, inTable, inHead)
		case "input":
			var ok bool
			var value string
			if value, ok = t.Attributes["type"]; !ok {
				return c.defaultInTableModeHandler(t)
			}

			if ok && !strings.EqualFold(value, "hidden") {
				return c.defaultInTableModeHandler(t)
			}

			c.insertHTMLElementForToken(t)
			spec.Pop(&c.stackOfOpenElements)
			// ack self closing
			return false, inTable, generalParseError
		case "form":
			if len(c.containedInStackOpenElements("template")) != 0 ||
				c.formElementPointer != nil {
				return false, inTable, generalParseError
			}

			elem := c.insertHTMLElementForToken(t)
			c.formElementPointer = elem
			spec.Pop(&c.stackOfOpenElements)
		}
	case endTagToken:
		switch t.TagName {
		case "table":
			mode := inTable
			c.stackContainsInScope("table", c.elementInTableScope, func() {
				spec.PopUntil(&c.stackOfOpenElements, "table")
				mode = c.resetInsertionMode()
			}, true)

			return false, mode, generalParseError
		case "body", "caption", "col", "colgroup", "html", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inTable, generalParseError
		case "template":
			return c.useRulesFor(t, inTable, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inTable, inBody)
	}
	return c.defaultInTableModeHandler(t)
}
func (c *HTMLTreeConstructor) inTableTextModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		if t.Data == "\u0000" {
			return false, inTableText, generalParseError
		}
		c.pendingTableCharacterTokens = append(c.pendingTableCharacterTokens, t)
		return false, inTableText, noError
	}

	isWhitespace := false
	for _, tok := range c.pendingTableCharacterTokens {
		if isASCIIWhitespace(int(tok.Data[0])) {
			isWhitespace = true
		}
	}

	for _, tok := range c.pendingTableCharacterTokens {
		if isWhitespace {
			c.defaultInTableModeHandler(tok)
		} else {
			c.insertCharacter(tok)
		}
	}
	return true, c.originalInsertionMode, noError
}
func (c *HTMLTreeConstructor) inCaptionModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) inColumnGroupModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) inTableBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "tr":
			c.clearStackBackToTableBody()
			c.insertHTMLElementForToken(t)
			return false, inRow, noError
		case "th", "td":
			c.clearStackBackToTableBody()
			tr := &Token{
				TokenType: startTagToken,
				TagName:   "tr",
			}
			c.insertHTMLElementForToken(tr)
			return true, inRow, generalParseError
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead":
			repro := true
			next := inTable
			c.stackContainsInScope("tbody", c.elementInTableScope, func() {
				c.stackContainsInScope("thead", c.elementInTableScope, func() {
					c.stackContainsInScope("tfoot", c.elementInTableScope, func() {
						repro = false
						next = inTableBody
						err = generalParseError
					}, false)
				}, false)
			}, false)

			if repro {
				c.clearStackBackToTableBody()
				spec.Pop(&c.stackOfOpenElements)
			}

			return repro, next, err
		}
	case endTagToken:
		switch t.TagName {
		case "tbody", "tfoot", "thead":
			next := inTableBody
			repro := false
			err = generalParseError
			c.stackContainsInScope(t.TagName, c.elementInTableScope, func() {
				c.clearStackBackToTableBody()
				spec.Pop(&c.stackOfOpenElements)
				next = inTable
				err = noError
				repro = true
			}, true)

			return repro, next, err
		case "table":
			repro := true
			next := inTable
			c.stackContainsInScope("tbody", c.elementInTableScope, func() {
				c.stackContainsInScope("thead", c.elementInTableScope, func() {
					c.stackContainsInScope("tfoot", c.elementInTableScope, func() {
						repro = false
						next = inTableBody
						err = generalParseError
					}, false)
				}, false)
			}, false)

			if repro {
				c.clearStackBackToTableBody()
				spec.Pop(&c.stackOfOpenElements)
			}

			return repro, next, err
		case "body", "caption", "col", "colgroup", "html", "td", "th", "tr":
			return false, inTableBody, generalParseError
		}
	}
	return c.useRulesFor(t, inTableBody, inTable)
}
func (c *HTMLTreeConstructor) inRowModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "th", "td":
			c.clearStackBackToTableRow()
			c.insertHTMLElementForToken(t)
			spec.Push(&c.activeFormattingElements, spec.ScopeMarker)
			return false, inCell, err
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr":
			next := inRow
			repro := false
			c.stackContainsInScope("tr", c.elementInTableScope, func() {
				c.clearStackBackToTableRow()
				spec.Pop(&c.stackOfOpenElements)
				next = inTableBody
				repro = true
			}, true)
			if !repro {
				err = generalParseError
			}

			return repro, next, err
		}
	case endTagToken:
		switch t.TagName {
		case "tr":
			next := inRow
			c.stackContainsInScope("tr", c.elementInTableScope, func() {
				c.clearStackBackToTableRow()
				spec.Pop(&c.stackOfOpenElements)
				next = inTableBody
			}, true)
			if next == inRow {
				err = generalParseError
			}

			return false, next, err
		case "table":
			next := inRow
			repro := false
			c.stackContainsInScope("tr", c.elementInTableScope, func() {
				c.clearStackBackToTableRow()
				spec.Pop(&c.stackOfOpenElements)
				next = inTableBody
				repro = true
			}, true)
			if !repro {
				err = generalParseError
			}

			return repro, next, err
		case "tbody", "tfoot", "thead":
			next := inRow
			repro := false
			c.stackContainsInScope(t.TagName, c.elementInTableScope, func() {
				c.stackContainsInScope("tr", c.elementInTableScope, func() {
					c.clearStackBackToTableRow()
					spec.Pop(&c.stackOfOpenElements)
					next = inTableBody
					repro = true
				}, false)
			}, true)

			return repro, next, err
		case "body", "caption", "col", "colgroup", "html", "td", "th":
			return false, inRow, generalParseError
		}
	}
	return c.useRulesFor(t, inRow, inTable)
}

func (c *HTMLTreeConstructor) closeCell() (bool, insertionMode, parseError) {
	err := noError
	c.generateImpliedEndTags([]webidl.DOMString{})
	cur := c.getCurrentNode().NodeName
	if cur != "td" && cur != "th" {
		err = generalParseError
	}
	for {
		node := spec.Pop(&c.stackOfOpenElements)
		if node.NodeName == "td" || node.NodeName == "th" {
			break
		}
	}
	c.clearListOfActiveFormattingElementsToLastMarker()
	return false, inRow, err
}

func (c *HTMLTreeConstructor) inCellModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "caption", "col", "colgroup", "tbody", "td", "tfoot", "th", "thead", "tr":
			repro := false
			err := noError
			c.stackContainsInScope("th", c.elementInTableScope, func() {
				c.stackContainsInScope("td", c.elementInTableScope, func() {
					repro = true
					err = generalParseError
					c.closeCell()
				}, true)
			}, true)
			return repro, inCell, err
		}
	case endTagToken:
		switch t.TagName {
		case "td", "th":
			err := noError
			next := inCell
			c.stackContainsInScope(t.TagName, c.elementInTableScope, func() {
				c.generateImpliedEndTags([]webidl.DOMString{})
				if c.getCurrentNode().NodeName != webidl.DOMString(t.TagName) {
					err = generalParseError
				}
				spec.PopUntil(&c.stackOfOpenElements, webidl.DOMString(t.TagName))
				spec.PopUntil(&c.activeFormattingElements, spec.ScopeMarker.NodeName)
				next = inRow
			}, true)

			return false, next, err
		case "body", "caption", "col", "colgroup", "html":
			return false, inCell, generalParseError
		case "table", "tbody", "tfoot", "thead", "tr":
			repro := false
			next := inCell
			err := noError
			c.stackContainsInScope(t.TagName, c.elementInTableScope, func() {
				_, next, err = c.closeCell()
				repro = true
			}, true)
			return repro, next, err
		}
	}
	return c.useRulesFor(t, inCell, inBody)
}

func (c *HTMLTreeConstructor) inSelectModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case characterToken:
		if t.Data == "\u0000" {
			err = generalParseError
		} else {
			c.insertCharacter(t)
		}
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		err = generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
		case "option":
			if c.getCurrentNode().NodeName == "option" {
				spec.Pop(&c.stackOfOpenElements)
			}
			c.insertHTMLElementForToken(t)
		case "optgroup":
			if c.getCurrentNode().NodeName == "option" {
				spec.Pop(&c.stackOfOpenElements)
			}

			if c.getCurrentNode().NodeName == "optgroup" {
				spec.Pop(&c.stackOfOpenElements)
			}
			c.insertHTMLElementForToken(t)
		case "select":
			err = generalParseError
			selectScope := false
			c.stackContainsInScope("select", c.elementInSelectScope, func() {
				selectScope = true
			}, true)
			if !selectScope {
				err = generalParseError
			} else {
				spec.PopUntil(&c.stackOfOpenElements, "select")
				return false, c.resetInsertionMode(), err
			}
		case "input", "keygen", "textarea":
		case "script", "template":

		}
	case endTagToken:
		switch t.TagName {
		case "optgroup":
		case "option":
		case "select":
		case "template":
		}
	case endOfFileToken:
		return c.useRulesFor(t, inSelect, inBody)
	default:
		err = generalParseError
	}
	return false, inSelect, err
}
func (c *HTMLTreeConstructor) inSelectInTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) inTemplateModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	case endOfFileToken:
		nodes := c.containedInStackOpenElements("template")
		if len(nodes) == 0 {
			return c.stopParsing(noError)
		}

		// parse error
		for {
			node := spec.Pop(&c.stackOfOpenElements)
			if node.NodeName == "template" {
				break
			}
		}
		c.clearListOfActiveFormattingElementsToLastMarker()
		c.stackOfInsertionModes = c.stackOfInsertionModes[:len(c.stackOfInsertionModes)-1]
		return true, c.resetInsertionMode(), generalParseError
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) afterBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
		if t.TagName == "html" {
			if c.createdBy == htmlFragmentParsingAlgorithm {
				return false, afterBody, generalParseError
			}
			return false, afterAfterBody, noError
		}
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) inFramesetModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) afterFramesetModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) afterAfterBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) afterAfterFramesetModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case endOfFileToken:
		return c.stopParsing(noError)
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}

//go:generate stringer -type=insertionMode
type insertionMode uint

const (
	initial insertionMode = iota
	beforeHTML
	beforeHead
	inHead
	inHeadNoScript
	afterHead
	inBody
	text
	inTable
	inTableText
	inCaption
	inColumnGroup
	inTableBody
	inRow
	inCell
	inSelect
	inSelectInTable
	inTemplate
	afterBody
	inFrameset
	afterFrameset
	afterAfterBody
	afterAfterFrameset
	stopParser
)

type treeConstructionModeHandler func(t *Token) (bool, insertionMode, parseError)

// ConstructTree constructs the HTML tree from the tokens that are emitted from the
// tokenizer.
func (c *HTMLTreeConstructor) ConstructTree() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s %s\n", err, de.Stack())
		}
		c.wg.Done()
	}()

	var (
		token     *Token
		nextMode  insertionMode
		reprocess bool
	)
	for token = range c.tokenChannel {
		if nextMode == stopParser {
			break
		}
		nextMode, reprocess = c.processToken(token, nextMode)
		for reprocess {
			nextMode, reprocess = c.processToken(token, nextMode)
		}
	}
}

var tagStateMappings = map[string][]insertionMode{
	"script":    {inHead},
	"noembed":   {inBody},
	"noscript":  {inHead, inBody},
	"textarea":  {inBody},
	"iframe":    {inBody},
	"noframes":  {inHead},
	"style":     {inHead},
	"title":     {inHead},
	"plaintext": {inBody},
	"xmp":       {inBody},
}

func specialTokenWrongState(token *Token, nextMode insertionMode) bool {
	if token.TokenType != startTagToken {
		return false
	}

	if nextMode != inHead && nextMode != inBody {
		return false
	}

	for tag, states := range tagStateMappings {
		if token.TagName != tag {
			continue
		}

		for _, state := range states {
			if nextMode == state {
				return true
			}
		}
	}

	return false
}

func (c *HTMLTreeConstructor) processToken(token *Token, nextMode insertionMode) (insertionMode, bool) {
	var (
		reprocess bool
		parseErr  parseError
	)
	fmt.Printf("token: %+v\n mode: %s\n", token, c.curInsertionMode)
	reprocess, c.curInsertionMode, parseErr = c.mappings[nextMode](token)
	if c.config[debug] == 0 {
		logError(parseErr)
	}
	// if we didn't consume the token, we don't want to check this state
	// only check if the token is in this special state when we are in our
	// consuming token state
	if !reprocess && specialTokenWrongState(token, c.curInsertionMode) {
		c.stateChannel <- dataState
	}
	return c.curInsertionMode, reprocess
}
