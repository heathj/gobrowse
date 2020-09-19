package parser

import (
	"browser/parser/spec"
	"browser/parser/webidl"
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
	config                                        htmlParserConfig
	HTMLDocument                                  *spec.HTMLDocument
	quirksMode                                    quirksMode
	fosterParenting                               bool
	scriptingEnabled                              bool
	originalInsertionMode                         insertionMode
	stackOfOpenElements, activeFormattingElements []*spec.Node
	headElementPointer                            *spec.Node
	formElementPointer                            *spec.Node
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
func NewHTMLTreeConstructor(c chan *Token, wg *sync.WaitGroup) *HTMLTreeConstructor {
	tr := HTMLTreeConstructor{
		tokenChannel: c,
		HTMLDocument: &spec.HTMLDocument{},
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
	return c.stackOfOpenElements[len(c.stackOfOpenElements)-1]
}

// Inserts a comment at a specific location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertCommentAt(t *Token) {
	//commentNode := spec\.NewComment(webidl.DOMString(t.Data))
	// I think if the window has multiple documents such as a webpage with an iframe
	// we will have to specific the document this comment needs to belong to.
	//commentNode.OwnerDocument = c.HTMLDocument.Document
	//handler(*commentNode.Node)

}

// Inserts a comment at the adjusted insertion location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertComment(t *Token) {
	//c.insertCommentAt(t, c.getAppropriatePlaceForInsertionDefault())
}

// https://html.spec.whatwg.org/multipage/parsing.html#appropriate-place-for-inserting-a-node
func (c *HTMLTreeConstructor) getAppropriatePlaceForInsertionDefault() *spec.Node {
	return c.getCurrentNode().LastChild
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

func (c *HTMLTreeConstructor) insertionTableSubSteps() *spec.Node {
	/*var lastTemplate *html.HTMLTemplateElement
	var lastTemplateI int
	for i, v := range c.stackOfOpenElements {
		switch v.(type) {
		case html.HTMLTemplateElement:
			lastTemplate = v.(*html.HTMLTemplateElement)
			lastTemplateI = i
		}
	}
	var lastTable *html.HTMLTableElement
	var lastTableI int
	for i, v := range c.stackOfOpenElements {
		switch v.(type) {
		case html.HTMLTableElement:
			lastTable = v.(*html.HTMLTableElement)
			lastTableI = i
		}
	}

	if lastTemplate == nil  (lastTable == nil || lastTemplateI > lastTableI) {
		// says template contents here
		return lastTemplate, func(e spec\.Node) { lastTemplate.AppendChild(e) }
	}

	if lastTable == nil {
		elem := c.stackOfOpenElements[len(c.stackOfOpenElements)-1]
		return elem, func(e spec\.Node) { elem.AppendChild(e) }
	}

	if lastTable.ParentNode != nil {
		// TODO: immediately before last table
		return lastTable, func(e spec\.Node) { lastTable.ParentNode.AppendChild(e) }
	}

	previousElement := c.stackOfOpenElements[lastTemplateI-1]
	return previousElement, func(e spec\.Node) { previousElement.AppendChild(e) }*/
	return nil

}

// https://html.spec.whatwg.org/multipage/parsing.html#appropriate-place-for-inserting-a-node
func (c *HTMLTreeConstructor) getAppropriatePlaceForInsertion(target *spec.Node) {
	/*var adjustedInsertionLocation spec\.Node = target
	var adjustedInsertionLocationHandler insertionHandler = func(e spec\.Node) { target.AppendChild(e) }
	if c.fosterParenting {
		switch target.(type) {
		case html.HTMLTableElement, html.HTMLTBodyElement, html.HTMLTFootElement, html.HTMLTHeadElement, html.HTMLTrElement:
			adjustedInsertionLocation, adjustedInsertionLocationHandler = c.insertionTableSubSteps()
		}
	}

	switch adjustedInsertionLocation.(type) {
	case html.HTMLTemplateElement:
		// TODO: template contents
	}

	return adjustedInsertionLocationHandler*/

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

//https://dom.spec.whatwg.org/#concept-create-element
func (c *HTMLTreeConstructor) createElement(document *spec.Document, localName string, ns namespace, optionals ...string) *spec.HTMLElement {
	/*prefix := ""
	if len(optionals) >= 1 {
		prefix = optionals[0]
	}
	is := ""
	if len(optionals) >= 2 {
		is = optionals[1]
	}

	var result *html.HTMLElement
	definition := c.lookUpCustomElementDefinition(document, ns, localName, is)
	if definition != nil  string(definition.name) != definition.localName {
		result = html.HTMLElement{Element: c.HTMLDocument.CreateElement(localName)}
		result.Element.Prefix = prefix
	} else if definition != nil {

	} else {

	}

	return result*/
	return nil

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
	loc := c.getAppropriatePlaceForInsertionDefault()
	if loc != nil && loc.ParentNode.NodeType == spec.DocumentNode {
		return
	}

	if loc.NodeType == spec.TextNode {
		loc.Text.CharacterData.Data += webidl.DOMString(t.Data)
	} else {
		tn := spec.NewTextNode(loc.OwnerDocument, t.Data)
		loc.ParentNode.AppendChild(tn)
	}

}

func (c *HTMLTreeConstructor) insertHTMLElementForToken(t *Token) *spec.Node {
	return c.insertForeignElementForToken(t, "html")
}

func (c *HTMLTreeConstructor) insertForeignElementForToken(t *Token, namespace webidl.DOMString) *spec.Node {
	loc := c.getAppropriatePlaceForInsertionDefault()
	elem := c.createElementForToken(t, namespace, loc)
	loc.ParentNode.AppendChild(elem)
	c.stackOfOpenElements = append(c.stackOfOpenElements, elem)
	return elem
}

func (c *HTMLTreeConstructor) useRulesFor(t *Token, returnState, expectedState insertionMode) (bool, insertionMode, parseError) {
	reprocess, nextstate, err := c.mappings[expectedState](t)

	// if the next state is the same as the expected state, this means that mode handler didn't
	// change the state. We should use the current return state.
	if nextstate == expectedState {
		return reprocess, returnState, err
	}
	return reprocess, expectedState, err
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
	for i := last; i >= 0 || elems < 3; i-- {
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

	c.activeFormattingElements = append(c.activeFormattingElements, elem)
}

func isSpecial(n *spec.Node) bool {
	switch n.NodeName {
	case "address", "applet", "area", "article", "aside", "base", "basefont", "bgsound", "blockquote", "body", "br", "button", "caption", "center", "col", "colgroup", "dd", "details", "dir", "div", "dl", "dt", "embed", "fieldset", "figcaption", "figure", "footer", "form", "frame", "frameset", "h1", "h2", "h3", "h4", "h5", "h6", "head", "header", "hgroup", "hr", "html", "iframe", "img", "input", "keygen", "li", "link", "listing", "main", "marquee", "menu", "meta", "nav", "noembed", "noframes", "noscript", "object", "ol", "p", "param", "plaintext", "pre", "script", "section", "select", "source", "style", "summary", "table", "tbody", "td", "template", "textarea", "tfoot", "th", "thead", "tr", "track", "ul", "wbr", "mi", "mo", "mn", "ms", "mtext", "annotation-xml", "foreignObject", "desc", "title":
		return true
	}
	return false
}

func (c *HTMLTreeConstructor) adoptionAgencyAlgorithm(t *Token) (bool, parseError) {
	var err parseError
	cur := c.getCurrentNode()
	if cur.TagName == webidl.DOMString(t.TagName) && spec.Contains(cur, &c.activeFormattingElements) == -1 {
		spec.Pop(&c.stackOfOpenElements)
		return false, noError
	}

	// outer loop
	var y, z, si, nif, nis int
	var formattingElement, furthestBlock *spec.Node
	for x := 0; x < 8; x++ {
		// 6
		for y = len(c.activeFormattingElements) - 1; y >= 0; y-- {
			if c.activeFormattingElements[y].TagName == webidl.DOMString(t.TagName) {
				formattingElement = c.activeFormattingElements[y]
			}

			if c.activeFormattingElements[y].NodeType == spec.ScopeMarkerNode {
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
		for z = y + 1; z < len(c.activeFormattingElements); z++ {
			if isSpecial(c.activeFormattingElements[z]) {
				furthestBlock = c.activeFormattingElements[z]
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
		for a := 1; ; a++ {
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
		ca.AppendChild(lastNode)

		// 16
		clone := formattingElement.CloneNode(true)
		clone.ParentNode = furthestBlock
		// 17
		for _, child := range furthestBlock.ChildNodes {
			clone.AppendChild(child)
		}
		// 18
		furthestBlock.AppendChild(clone)
		// 19
		f := spec.Contains(formattingElement, &c.activeFormattingElements)
		if f != -1 {
			spec.Remove(f, &c.activeFormattingElements)
			b := spec.Contains(bm, &c.activeFormattingElements)
			if b != 1 {
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
					c.stackOfOpenElements = append(c.stackOfOpenElements, clone)
				} else {
					c.stackOfOpenElements[b+1] = clone
				}
			}
		}
	}

	return false, err
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
	// 5. Let entry be the entry one earlier than entry in the list of active formatting elements.
	for ; i >= 0; i-- {
		// 6. If entry is neither a marker nor an element that is also in the stack of open elements,
		// go to the step labeled rewind.
		doesContain := spec.Contains(c.activeFormattingElements[i], &c.stackOfOpenElements)
		if c.activeFormattingElements[i].NodeType == spec.ScopeMarkerNode || doesContain != -1 {
			break
		}
	}

	// 7. Advance: Let entry be the element one later than entry in the list of active formatting
	// elements.
	for ; i < len(c.activeFormattingElements)-1; i++ {
		// 8. Create: Insert an HTML element for the token for which the element entry was created,
		// to obtain new element.
		loc := c.getAppropriatePlaceForInsertionDefault()
		elem := c.activeFormattingElements[i+1].CloneNode(true)
		loc.ParentNode.AppendChild(elem)
		c.stackOfOpenElements = append(c.stackOfOpenElements, elem)

		// 9. Replace the entry for entry in the list with an entry for new element.
		c.activeFormattingElements[i+1] = elem

		// 10. If the entry for new element in the list of active formatting elements is not the last
		// entry in the list, return to the step labeled advance.
	}

}

func (c *HTMLTreeConstructor) clearStackBackToTable() {
	/*for {
		switch c.CurrentNode().elemType {
		case tableElement, templateElement, htmlElement:
			return
		default:
			c.PopOpenElements()
		}
	}*/
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
	case docTypeToken:
		if t.TagName != "html" ||
			t.PublicIdentifier != missing ||
			(t.SystemIdentifier != missing &&
				t.SystemIdentifier != "about:legacy-compat") {
			//TODO: just says this was a parse error?
			err = generalParseError
		}

		doctype := spec.NewDocTypeNode(t.TagName, t.PublicIdentifier, t.SystemIdentifier)
		c.HTMLDocument.AppendChild(doctype)
		c.HTMLDocument.Doctype = doctype

		if c.isForceQuirks(t) {
			c.quirksMode = quirks
		} else if c.isLimitedQuirks(t) {
			c.quirksMode = limitedQuirks
		} else {
			c.quirksMode = noQuirks
		}

		return false, beforeHTML, err
	default:
		/**/
	}
	return true, beforeHTML, err
}

func (c *HTMLTreeConstructor) defaultBeforeHTMLModeHandler(t *Token) (bool, insertionMode, parseError) {
	//TODO
	//	elem := c.CreateHTMLElement()
	//	c.PushOpenElements(elem)
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
			c.stackOfOpenElements = append(c.stackOfOpenElements, elem)
			// handle navigation of a browsing context
		}
		return false, beforeHead, noError
	case endTagToken:
		switch t.TagName {
		case "head", "body", "html", "br":
			return c.defaultBeforeHTMLModeHandler(t)
		default:
			return false, beforeHTML, generalParseError
		}
	default:
		return c.defaultBeforeHTMLModeHandler(t)
	}
	return false, initial, noError
}

func (c *HTMLTreeConstructor) defaultBeforeHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	//TODO: Insert an HTML element for a "head" start tag token with no attributes.
	/*c.WriteHTMLElement(t)
	c.WriteLatestHeadElement()*/
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

func (c *HTMLTreeConstructor) defaultInHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	return false, initial, noError
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
		case "noscript":
			if c.scriptingEnabled {

			} else {
				//c.WriteHTMLElement(t)
				return false, inHeadNoScript, noError
			}
		case "noframes", "style":
		case "script":
		case "template":
		case "head":
			return false, inHead, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "head":
			c.stackOfOpenElements = c.stackOfOpenElements[:len(c.stackOfOpenElements)-1]
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
	/*c.WriteHTMLElement(Token{
		TokenType: startTagToken,
		TagName:   "body",
	})*/
	return true, inBody, noError
}
func (c *HTMLTreeConstructor) afterHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			//c.WriteCharacter(t)
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

func (c *HTMLTreeConstructor) defaultInBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	return false, inBody, noError
}

func (c *HTMLTreeConstructor) inBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case characterToken:
		if t.Data == "\u0000" {
			return false, inBody, generalParseError
		}

		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.reconstructActiveFormattingElements()
			c.insertCharacter(t)
			return false, inBody, noError
		default:
			c.reconstructActiveFormattingElements()
			c.insertCharacter(t)
			c.frameset = framesetOK
			return false, inBody, noError
		}
	case commentToken:
		c.insertComment(t)
		return false, inBody, noError
	case docTypeToken:
		return false, inBody, generalParseError
	case startTagToken:

		switch t.TagName {
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style", "template", "title":
			return c.useRulesFor(t, inBody, inHead)
		case "body":
		case "frameset":
		case "address", "article", "aside", "blockquote", "center", "details", "dialog", "dir", "div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "main", "menu", "nav", "ol", "p", "section", "summary", "ul":
		case "h1", "h2", "h3", "h4", "h5", "h6":
		case "pre", "listing":
		case "form":
		case "li":
		case "dd", "dt":
		case "plaintext":
		case "button":
		case "a":
		case "b", "big", "code", "em", "font", "i", "s", "small", "strike", "strong", "tt", "u":
			c.reconstructActiveFormattingElements()
			elem := c.insertHTMLElementForToken(t)
			c.activeFormattingElements = append(c.activeFormattingElements, elem)
		case "nobr":
		case "applet", "marquee", "object":
		case "table":
		case "area", "br", "embed", "img", "keygen", "wbr":
		case "input":
		case "param", "source", "track":
		case "hr":
		case "image":
		case "textarea":
		case "xmp":
		case "iframe":
		case "noembed":
		case "noscript":
		case "select":
		case "optgroup", "option":
		case "rb", "rtc":
		case "rp", "rt":
		case "math":
		case "svg":
		case "caption", "col", "colgroup", "frame", "head", "tbody", "td", "tfoot", "th", "thead", "tr":
		default:
		}
		return false, inBody, err
	case endTagToken:
		switch t.TagName {
		case "template":
		case "body":
		case "html":
		case "address", "article", "aside", "blockquote", "button", "center", "details", "dialog", "dir", "div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "listing", "main", "menu", "nav", "ol", "pre", "section", "summary", "ul":
		case "form":
		case "p":
		case "li":
		case "dd", "dt":
		case "h1", "h2", "h3", "h4", "h5", "h6":
		case "sarcasm":
		case "a", "b", "big", "code", "em", "font", "i", "nobr", "s", "small", "strike", "strong", "tt", "u":
			var shouldDefault bool
			shouldDefault, err = c.adoptionAgencyAlgorithm(t)
			if shouldDefault {
				a, b, _ := c.defaultInBodyModeHandler(t)
				return a, b, err
			}
			return false, inBody, err
		case "applet", "marquee", "object":
		case "br":
		default:
			return c.defaultInBodyModeHandler(t)
		}

		return false, inBody, err
	case endOfFileToken:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) textModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		//c.WriteCharacter(t)
		return false, text, noError
	case endOfFileToken:
		return true, c.originalInsertionMode, generalParseError
	case endTagToken:
		switch t.TagName {
		case "script":
			return false, c.originalInsertionMode, noError
		default:
			//c.PopOpenElements()
			return false, c.originalInsertionMode, noError
		}
	}
	return false, text, noError
}

func (c *HTMLTreeConstructor) defaultInTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	return false, inTable, noError
}
func (c *HTMLTreeConstructor) inTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		/*switch c.CurrentNode().elemType {
		case tableElement, tbodyElement, tfootElement, theadElement, trElement:
			c.originalInsertionMode = inTable
			return true, inTableText, noError
		}*/
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		return false, inTable, generalParseError
	case startTagToken:
		switch t.TagName {
		case "caption":
			c.clearStackBackToTable()
			/*c.WriteMarker()
			c.WriteHTMLElement(t)*/
			return false, inColumnGroup, noError
		case "colgroup":
			c.clearStackBackToTable()
			//c.WriteHTMLElement(t)
			return false, inColumnGroup, noError
		case "col":
			c.clearStackBackToTable()
			/*c.WriteHTMLElement(Token{
				TagName:   "colgroup",
				TokenType: startTagToken,
			})*/
			return true, inColumnGroup, noError
		case "tbody", "tfoot", "thead":
			c.clearStackBackToTable()
			//c.WriteHTMLElement(t)
			return false, inTableBody, noError
		case "td", "th", "tr":
			c.clearStackBackToTable()
			/*c.WriteHTMLElement(Token{
				TokenType: startTagToken,
				TagName:   "tbody",
			})*/
			return true, inTableBody, noError
		case "table":
			/*if c.openElementsInTableScope(tableElement) {
				return false, inTable, noError
			}*/

			c.clearStackBackToTable()
			/*c.PopOpenElements()
			nextMode := c.resetInsertionMode()*/
			nextMode := beforeHTML
			return true, nextMode, noError
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

			/*c.WriteHTMLElement(t)
			c.PopOpenElements()*/
			//TODO self closing
			return false, inTable, generalParseError
		case "form":
			/*if c.SearchOpenElements(templateElement) || c.formPointer != nil {
				return false, inTable, generalParseError
			}

			c.WriteHTMLElement(t)*/
			//			c.formPointer =
			//c.PopOpenElements()
		}
	case endTagToken:
		switch t.TagName {
		case "table":
		case "body", "caption", "col", "colgroup", "html", "tbody", "td", "tfoot", "th", "thead", "tr":
		case "template":
		}
	}
	c.useRulesFor(t, inTable, inBody)
	return false, initial, generalParseError
}
func (c *HTMLTreeConstructor) inTableTextModeHandler(t *Token) (bool, insertionMode, parseError) {
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
func (c *HTMLTreeConstructor) inRowModeHandler(t *Token) (bool, insertionMode, parseError) {
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
func (c *HTMLTreeConstructor) inCellModeHandler(t *Token) (bool, insertionMode, parseError) {
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
func (c *HTMLTreeConstructor) inSelectModeHandler(t *Token) (bool, insertionMode, parseError) {
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
	case startTagToken:
	case endTagToken:
	default:
	}
	return false, initial, noError
}

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
)

type treeConstructionModeHandler func(t *Token) (bool, insertionMode, parseError)

// ConstructTree constructs the HTML tree from the tokens that are emitted from the
// tokenizer.
func (c *HTMLTreeConstructor) ConstructTree() {
	c.wg.Add(1)
	defer c.wg.Done()
	var (
		token           *Token
		nextModeHandler treeConstructionModeHandler
		nextMode        insertionMode
		parseErr        parseError
		reprocess       bool
	)
	for {
		token = <-c.tokenChannel
		nextModeHandler = c.mappings[nextMode]
		reprocess, nextMode, parseErr = nextModeHandler(token)
		if c.config[debug] == 0 {
			logError(parseErr)
		}

		for {
			if !reprocess {
				break
			}

			nextModeHandler = c.mappings[nextMode]
			reprocess, nextMode, parseErr = nextModeHandler(token)
			if c.config[debug] == 0 {
				logError(parseErr)
			}
		}
	}
}
