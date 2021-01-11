package parser

import (
	"browser/parser/spec"
	"browser/parser/webidl"
	"fmt"
	de "runtime/debug"
	"strings"
	"sync"
)

type quirksMode uint

const (
	noQuirks quirksMode = iota
	quirks
	limitedQuirks
)

type createdByOrigin uint

const (
	htmlParsingNormal createdByOrigin = iota
	htmlFragmentParsingAlgorithm
)

type frameset uint

const (
	framesetOK frameset = iota
	framesetNotOK
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
	context                                       *spec.Node
	originalInsertionMode                         insertionMode
	stackOfOpenElements, activeFormattingElements spec.NodeList
	stackOfTemplateInsertionModes                 []insertionMode
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
		initial:             c.initialModeHandler,
		beforeHTML:          c.beforeHTMLModeHandler,
		beforeHead:          c.beforeHeadModeHandler,
		inHead:              c.inHeadModeHandler,
		inHeadNoScript:      c.inHeadNoScriptModeHandler,
		afterHead:           c.afterHeadModeHandler,
		inBody:              c.inBodyModeHandler,
		inBodyPeekNextToken: c.inBodyPeekNextToken,
		text:                c.textModeHandler,
		inTable:             c.inTableModeHandler,
		inTableText:         c.inTableTextModeHandler,
		inCaption:           c.inCaptionModeHandler,
		inColumnGroup:       c.inColumnGroupModeHandler,
		inTableBody:         c.inTableBodyModeHandler,
		inRow:               c.inRowModeHandler,
		inCell:              c.inCellModeHandler,
		inSelect:            c.inSelectModeHandler,
		inSelectInTable:     c.inSelectInTableModeHandler,
		inTemplate:          c.inTemplateModeHandler,
		afterBody:           c.afterBodyModeHandler,
		inFrameset:          c.inFramesetModeHandler,
		afterFrameset:       c.afterFramesetModeHandler,
		afterAfterBody:      c.afterAfterBodyModeHandler,
		afterAfterFrameset:  c.afterAfterFramesetModeHandler,
	}
}

func (c *HTMLTreeConstructor) getCurrentNode() *spec.Node {
	if len(c.stackOfOpenElements) == 0 {
		return nil
	}
	return c.stackOfOpenElements[len(c.stackOfOpenElements)-1]
}

func (c *HTMLTreeConstructor) getAdjustedCurrentNode() *spec.Node {
	if c.createdBy != htmlFragmentParsingAlgorithm {
		return c.getCurrentNode()
	}

	return c.context
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

func targetInTable(name string) bool {
	switch name {
	case "table", "tbody", "tfoot", "thead", "tr":
		return true
	}
	return false
}

func (c *HTMLTreeConstructor) adjustMathMLAttrs(t *Token) {
	if val, ok := t.Attributes["definitionurl"]; ok {
		delete(t.Attributes, "definitionurl")
		t.Attributes["definitionURL"] = val
	}
}

var svgAttrTable = map[string]string{
	"attributename":       "attributeName",
	"attributetype":       "attributeType",
	"basefrequency":       "baseFrequency",
	"baseprofile":         "baseProfile",
	"calcmode":            "calcMode",
	"clippathunits":       "clipPathUnits",
	"diffuseconstant":     "diffuseConstant`",
	"edgemode":            "edgeMode",
	"filterunits":         "filterUnits",
	"glyphref":            "glyphRef",
	"gradienttransform":   "gradientTransform",
	"gradientunits":       "gradientUnits",
	"kernelmatrix":        "kernelMatrix",
	"kernelunitlength":    "kernelUnitLength",
	"keypoints":           "keyPoints",
	"keysplines":          "keySplines",
	"keytimes":            "keyTimes",
	"lengthadjust":        "lengthAdjust",
	"limitingconeangle":   "limitingConeAngle",
	"markerheight":        "markerHeight",
	"markerunits":         "markerUnits",
	"markerwidth":         "markerWidth",
	"maskcontentunits":    "maskContentUnits",
	"maskunits":           "maskUnits",
	"numoctaves":          "numOctaves",
	"pathlength":          "pathLength",
	"patterncontentunits": "patternContentUnits",
	"patterntransform":    "patternTransform",
	"patternunits":        "patternUnits",
	"pointsatx":           "pointsAtX",
	"pointsaty":           "pointsAtY",
	"pointsatz":           "pointsAtZ",
	"preservealpha":       "preserveAlpha",
	"preserveaspectratio": "preserveAspectRatio",
	"primitiveunits":      "primitiveUnits",
	"refx":                "refX",
	"refy":                "refY",
	"repeatcount":         "repeatCount",
	"repeatdur":           "repeatDur",
	"requiredextensions":  "requiredExtensions",
	"requiredfeatures":    "requiredFeatures",
	"specularconstant":    "specularConstant",
	"specularexponent":    "specularExponent",
	"spreadmethod":        "spreadMethod",
	"startoffset":         "startOffset",
	"stddeviation":        "stdDeviation",
	"stitchtiles":         "stitchTiles",
	"surfacescale":        "surfaceScale",
	"systemlanguage":      "systemLanguage",
	"tablevalues":         "tableValues",
	"targetx":             "targetX",
	"targety":             "targetY",
	"textlength":          "textLength",
	"viewbox":             "viewBox",
	"viewtarget":          "viewTarget",
	"xchannelselector":    "xChannelSelector",
	"ychannelselector":    "yChannelSelector",
	"zoomandpan":          "zoomAndPan",
}

func (c *HTMLTreeConstructor) adjustSVGAttrs(t *Token) {
	for k := range t.Attributes {
		if val, ok := svgAttrTable[k]; ok {
			t.Attributes[k] = val
		}
	}
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
	if c.fosterParenting && targetInTable(string(target.NodeName)) {
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
	} else {
		ail.node = target
		ail.insert = func(n *spec.Node) { target.AppendChild(n) }
	}

	return ail
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
func (c *HTMLTreeConstructor) lookUpCustomElementDefinition(document *spec.Node, ns spec.Namespace, localName, is webidl.DOMString) *CustomElementDefinition {

	// browsing context
	// custom element registry
	return nil
}

func (c *HTMLTreeConstructor) clearListOfActiveFormattingElementsToLastMarker() {
	for {
		node := c.activeFormattingElements.Pop()
		if node == nil || node.NodeType == spec.ScopeMarkerNode {
			return
		}
	}
}

func (c *HTMLTreeConstructor) resetInsertionMode() insertionMode {
	return c.resetInsertionModeWithContext(nil)
}

func (c *HTMLTreeConstructor) resetInsertionModeWithContext(context *spec.Node) insertionMode {
	last := false
	lastID := len(c.stackOfOpenElements) - 1
	node := c.stackOfOpenElements[lastID]
	j := lastID
	for {
		if j == 0 {
			last = true
			if c.createdBy == htmlFragmentParsingAlgorithm &&
				context != nil {
				node = context
			}
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
			return c.stackOfTemplateInsertionModes[len(c.stackOfTemplateInsertionModes)-1]
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
		c.stackOfOpenElements.Pop()
	}

	return false, stopParser, err
}

// createElementForToken creates an element from a token with the provided
// namespace and parent element.
// https://html.spec.whatwg.org/multipage/parsing.html#create-an-element-for-the-token
func (c *HTMLTreeConstructor) createElementForToken(t *Token, ns spec.Namespace, ip *spec.Node) *spec.Node {
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

	tn := spec.NewTextNode(il.node.OwnerDocument, t.Data)
	il.insert(tn)
	if tn.PreviousSibling != nil && tn.PreviousSibling.NodeType == spec.TextNode {
		tn.PreviousSibling.Text.CharacterData.Data += webidl.DOMString(t.Data)
		il.node.RemoveChild(tn)
	} else if tn.NextSibling != nil && tn.NextSibling.NodeType == spec.TextNode {
		tn.NextSibling.Text.CharacterData.Data = webidl.DOMString(t.Data) + tn.NextSibling.Text.CharacterData.Data
		il.node.RemoveChild(tn)
	}
}

func (c *HTMLTreeConstructor) insertHTMLElementForToken(t *Token) *spec.Node {
	return c.insertForeignElementForToken(t, spec.Htmlns)
}

func (c *HTMLTreeConstructor) insertForeignElementForToken(t *Token, namespace spec.Namespace) *spec.Node {
	il := c.getAppropriatePlaceForInsertion(nil)
	elem := c.createElementForToken(t, namespace, il.node)
	il.insert(elem)
	c.stackOfOpenElements.Push(elem)
	return elem
}

func (c *HTMLTreeConstructor) useRulesFor(t *Token, mode insertionMode) (bool, insertionMode, parseError) {
	reprocess, nextMode := c.processToken(t, mode)

	// if the mode didn't change return to where we came from.
	if nextMode == mode {
		return reprocess, c.curInsertionMode, noError
	}
	return reprocess, nextMode, noError
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

	c.activeFormattingElements.Push(elem)
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
			c.stackOfOpenElements.Pop()
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
	c.stackOfOpenElements.PopUntil("p")
}

func (c *HTMLTreeConstructor) adoptionAgencyAlgorithm(t *Token) (bool, parseError) {
	var err parseError
	cur := c.getCurrentNode()
	if cur.NodeName == webidl.DOMString(t.TagName) && c.activeFormattingElements.Contains(cur) == -1 {
		c.stackOfOpenElements.Pop()
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
		si = c.stackOfOpenElements.Contains(formattingElement)
		if si == -1 {
			// parse error
			c.activeFormattingElements.Remove(y)
			return false, noError
		}

		// 8
		if !c.stackOfOpenElements.ContainsElementInScope(formattingElement.NodeName) {
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
					c.stackOfOpenElements.Pop()
					c.activeFormattingElements.Remove(y)
					return false, noError
				}
				c.stackOfOpenElements.Pop()
			}
		}

		// 12
		ca := c.stackOfOpenElements[si-1]
		// 13
		bm := y

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
				c.activeFormattingElements.Remove(c.activeFormattingElements.Contains(node))
			}

			// 14.6
			if c.activeFormattingElements.Contains(node) == -1 {
				c.stackOfOpenElements.Remove(c.stackOfOpenElements.Contains(node))
				continue
			}

			// 14.7
			clone := node.CloneNode(false)
			nif = c.activeFormattingElements.Contains(node)
			if nif != -1 {
				c.activeFormattingElements[nif] = clone
			}
			nis = c.stackOfOpenElements.Contains(node)
			if nis != -1 {
				c.stackOfOpenElements[nis] = clone
			}
			// need to replace the node with the clone so that the references match
			// up when replacing the bookmark below and append a child of the last node.
			node = clone

			// 14.8
			if lastNode == furthestBlock {
				bm = nif + 1
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
		il := c.getAppropriatePlaceForInsertion(ca)
		il.insert(lastNode)

		// 16
		clone := formattingElement.CloneNode(false)
		clone.ParentNode = furthestBlock
		// 17
		for len(furthestBlock.ChildNodes) > 0 {
			// same as above, here. This step does NOT explicitly say to remove the children elements
			// and move them, but the algorithm wouldn't really work otherwise.
			removed := furthestBlock.ChildNodes.Remove(0)
			if removed == nil {
				break
			}
			clone.AppendChild(removed)
		}
		// 18
		furthestBlock.AppendChild(clone)
		// 19
		f := c.activeFormattingElements.Contains(formattingElement)
		if f != -1 {
			c.activeFormattingElements.Remove(f)
			// shifting the bookmark after removing the element above. we only shift
			// though if the bookmark was later in the list. if f above was after the bookmark
			// position, the position wouldn't change:
			// [1, 2, f, bm] -> [1, 2, bm] (position changed)
			// [bm, 1, 2, f] -> [bm, 1, 2] (no position changed)
			if f < bm {
				bm--
			}
			c.activeFormattingElements.WedgeIn(bm, clone)
		}

		//20
		f = c.stackOfOpenElements.Contains(formattingElement)
		if f != -1 {
			c.stackOfOpenElements.Remove(f)
			b := c.stackOfOpenElements.Contains(furthestBlock)
			if b != -1 {
				if b+1 == len(c.stackOfOpenElements) {
					c.stackOfOpenElements.Push(clone)
				} else {
					c.stackOfOpenElements.WedgeIn(b+1, clone)
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
	elem := c.activeFormattingElements[i].CloneNode(false)
	il.insert(elem)
	c.stackOfOpenElements.Push(elem)

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
	doesContain := c.stackOfOpenElements.Contains(lafe)
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
			doesContain = c.stackOfOpenElements.Contains(c.activeFormattingElements[i])
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
		c.stackOfOpenElements.Pop()
	}
}

func (c *HTMLTreeConstructor) clearStackBackToTableRow() {
	for {
		switch c.getCurrentNode().NodeName {
		case "tr", "template", "html":
			return
		}
		c.stackOfOpenElements.Pop()
	}
}

func (c *HTMLTreeConstructor) clearStackBackToTableBody() {
	for {
		switch c.getCurrentNode().NodeName {
		case "tbody", "tfoot", "thead", "template", "html":
			return
		}
		c.stackOfOpenElements.Pop()
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
		}

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
	// TODO:only if not an iframe src document
	c.quirksMode = quirks
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
	n := spec.NewDOMElement(c.HTMLDocument.Node, "html", spec.Htmlns)
	n.OwnerDocument = c.HTMLDocument.Node
	c.HTMLDocument.AppendChild(n)
	c.stackOfOpenElements.Push(n)

	// TODO: application cache algo
	return true, beforeHead, noError
}

// https://html.spec.whatwg.org/multipage/parsing.html#the-before-html-insertion-mode
func (c *HTMLTreeConstructor) beforeHTMLModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case docTypeToken:
		return false, beforeHTML, generalParseError
	case commentToken:
		il := &insertionLocation{
			node: c.HTMLDocument.Node,
			insert: func(n *spec.Node) {
				c.HTMLDocument.Node.AppendChild(n)
			},
		}
		c.insertCommentAt(t, il)
		return false, beforeHTML, noError
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, beforeHTML, noError
		}
	case startTagToken:
		if t.TagName == "html" {
			elem := c.createElementForToken(t, spec.Htmlns, c.HTMLDocument.Node)
			c.HTMLDocument.AppendChild(elem)
			c.stackOfOpenElements.Push(elem)
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
			return c.useRulesFor(t, inBody)
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

func (c *HTMLTreeConstructor) genericRCDATAElementParsingAlgorithm(t *Token) (bool, insertionMode, parseError) {
	c.insertHTMLElementForToken(t)
	c.originalInsertionMode = c.curInsertionMode
	c.switchTokenizerState(t, rcDataState)
	return false, text, noError
}

func (c *HTMLTreeConstructor) genericRawTextElementParsingAlgorithm(t *Token) (bool, insertionMode, parseError) {
	c.insertHTMLElementForToken(t)
	c.originalInsertionMode = c.curInsertionMode
	c.switchTokenizerState(t, rawTextState)
	return false, text, noError
}

func (c *HTMLTreeConstructor) defaultInHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	c.stackOfOpenElements.Pop()
	return true, afterHead, noError
}
func (c *HTMLTreeConstructor) inHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.insertCharacter(t)
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
			return c.useRulesFor(t, inBody)
		case "base", "basefont", "bgsound", "link":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			return false, inHead, noError
			//TODO: acknowledge the self closing flag?
		case "meta":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			//TODO: acknowledge the self closing flag?
			//TODO: char encoding settings
			return false, inHead, noError
		case "title":
			return c.genericRCDATAElementParsingAlgorithm(t)
		case "noscript":
			if c.scriptingEnabled {
				return c.genericRawTextElementParsingAlgorithm(t)
			}
			c.insertHTMLElementForToken(t)
			return false, inHeadNoScript, noError
		case "noframes", "style":
			return c.genericRawTextElementParsingAlgorithm(t)
		case "script":
			il := c.getAppropriatePlaceForInsertion(nil)
			elem := c.createElementForToken(t, spec.Htmlns, il.node)
			elem.ParserDocument = c.HTMLDocument
			elem.NonBlocking = false
			il.insert(elem)
			c.stackOfOpenElements.Push(elem)
			c.stateChannel <- scriptDataState
			t.Special = false
			c.originalInsertionMode = c.curInsertionMode
			return false, text, noError
		case "template":
		case "head":
			return false, inHead, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "head":
			c.stackOfOpenElements.Pop()
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
	c.stackOfOpenElements.Pop()
	return true, inHead, generalParseError
}
func (c *HTMLTreeConstructor) inHeadNoScriptModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return c.useRulesFor(t, inHead)
		}
	case commentToken:
		return c.useRulesFor(t, inHead)
	case docTypeToken:
		return false, inHeadNoScript, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "basefont", "bgsound", "link", "meta", "noframe", "style":
			return c.useRulesFor(t, inHead)
		case "head", "noscript":
			return false, inHeadNoScript, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "noscript":
			c.stackOfOpenElements.Pop()
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
	err := noError
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.insertCharacter(t)
			return false, afterHead, err
		}
	case commentToken:
		c.insertComment(t)
		return false, afterHead, err
	case docTypeToken:
		return false, afterHead, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "body":
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inBody, noError
		case "frameset":
			c.insertHTMLElementForToken(t)
			return false, inFrameset, noError
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style", "template", "title":
			err = generalParseError
			c.stackOfOpenElements.Push(c.headElementPointer)
			repro, nextState, err := c.useRulesFor(t, inHead)
			c.stackOfOpenElements.Remove(c.stackOfOpenElements.Contains(c.headElementPointer))
			return repro, nextState, err
		case "head":
			return false, afterHead, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "template":
			return c.useRulesFor(t, inHead)
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
				popped := c.stackOfOpenElements.Pop()
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

func (c *HTMLTreeConstructor) inBodyPeekNextToken(t *Token) (bool, insertionMode, parseError) {
	ret := inBody
	if c.getCurrentNode().NodeName == "textarea" {
		ret = text
	}
	if t.TokenType == characterToken && t.Data == "\u000A" {
		return false, ret, noError
	}
	return true, ret, noError
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
		case "html":
			if len(c.containedInStackOpenElements("template")) > 0 {
				return false, inBody, generalParseError
			}

			for k, v := range t.Attributes {
				if _, ok := c.stackOfOpenElements[0].Attributes.Attrs[k]; !ok {
					c.stackOfOpenElements[0].Attributes.Attrs[k] = v
				}
			}
			return false, inBody, generalParseError
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style",
			"template", "title":
			return c.useRulesFor(t, inHead)
		case "body":
			err = generalParseError
			if len(c.stackOfOpenElements) <= 1 ||
				c.stackOfOpenElements[1].NodeName != "body" ||
				len(c.containedInStackOpenElements("template")) != 0 {
				return false, inBody, err
			}

			c.frameset = framesetNotOK
			for k, v := range t.Attributes {
				if _, ok := c.stackOfOpenElements[1].Attributes.Attrs[k]; !ok {
					c.stackOfOpenElements[1].Attributes.Attrs[k] = v
				}
			}
		case "frameset":
			err = generalParseError
			if len(c.stackOfOpenElements) <= 1 ||
				c.stackOfOpenElements[1].NodeName != "body" {
				return false, inBody, err
			}
			if c.frameset == framesetNotOK {
				return false, inBody, err
			}
			c.stackOfOpenElements[1].ParentNode.RemoveChild(c.stackOfOpenElements[1])
			for c.getCurrentNode().NodeName != "html" {
				c.stackOfOpenElements.Pop()
			}
			c.insertHTMLElementForToken(t)
			return false, inFrameset, err
		case "address", "article", "aside", "blockquote", "center", "details", "dialog", "dir",
			"div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "main",
			"menu", "nav", "ol", "p", "section", "summary", "ul":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
		case "h1", "h2", "h3", "h4", "h5", "h6":
			err = noError
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			switch c.getCurrentNode().NodeName {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				err = generalParseError
				c.stackOfOpenElements.Pop()
			}

			c.insertHTMLElementForToken(t)
			return false, inBody, err
		case "pre", "listing":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inBodyPeekNextToken, noError
		case "form":
			noTemp := len(c.containedInStackOpenElements("template")) == 0
			if c.formElementPointer != nil && noTemp {
				return false, inBody, generalParseError
			}

			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			elem := c.insertHTMLElementForToken(t)
			if noTemp {
				c.formElementPointer = elem
			}
		case "li":
			done := func() (bool, insertionMode, parseError) {
				if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
					c.closePElement()
				}
				c.insertHTMLElementForToken(t)
				return false, inBody, err
			}
			node := c.getCurrentNode()
			c.frameset = framesetNotOK
			i := 1
			for node.NodeName != "li" {
				if isSpecial(node) {
					switch node.NodeName {
					case "address", "div", "p":
						node = c.stackOfOpenElements[len(c.stackOfOpenElements)-1-i]
						i++
						continue
					default:
						return done()
					}
				}

			}

			c.generateImpliedEndTags([]webidl.DOMString{"li"})
			if c.getCurrentNode().NodeName != "li" {
				err = generalParseError
			}
			c.stackOfOpenElements.PopUntil("li")

			// done
			return done()
		case "dd", "dt":
			c.frameset = framesetNotOK
			var node *spec.Node
			for i := len(c.stackOfOpenElements) - 1; i >= 0; i-- {
				node = c.stackOfOpenElements[i]
				if node.NodeName == "dd" {
					c.generateImpliedEndTags([]webidl.DOMString{"dd"})
					if c.getCurrentNode().NodeName != "dd" {
						err = generalParseError
					}
					c.stackOfOpenElements.PopUntil("dd")
					break
				}

				if node.NodeName == "dt" {
					c.generateImpliedEndTags([]webidl.DOMString{"dt"})
					if c.getCurrentNode().NodeName != "dt" {
						err = generalParseError
					}
					c.stackOfOpenElements.PopUntil("dt")
					break
				}

				if isSpecial(node) {
					switch node.NodeName {
					case "address", "div", "p":
					default:
						break
					}
				}
			}

			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			return false, inBody, err
		case "plaintext":
			if c.stackOfOpenElements.ContainsElementInButtonScope("button") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			c.switchTokenizerState(t, plaintextState)
		case "button":
			if c.stackOfOpenElements.ContainsElementInScope("button") {
				c.generateImpliedEndTags([]webidl.DOMString{})
				c.stackOfOpenElements.PopUntil("button")
			}
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
		case "a":
			var node *spec.Node
			for i := len(c.activeFormattingElements) - 1; i >= 0; i-- {
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
					c.activeFormattingElements.Remove(c.activeFormattingElements.Contains(node))
					c.stackOfOpenElements.Remove(c.stackOfOpenElements.Contains(node))
					break
				}
			}

			c.reconstructActiveFormattingElements()
			elem := c.insertHTMLElementForToken(t)
			c.pushActiveFormattingElements(elem)
		case "b", "big", "code", "em", "font", "i", "s", "small", "strike", "strong", "tt", "u":
			c.reconstructActiveFormattingElements()
			elem := c.insertHTMLElementForToken(t)
			c.activeFormattingElements.Push(elem)
		case "nobr":
			c.reconstructActiveFormattingElements()
			if c.stackOfOpenElements.ContainsElementInScope("nobr") {
				err = generalParseError
				c.adoptionAgencyAlgorithm(t)
				c.reconstructActiveFormattingElements()
			}
			elem := c.insertHTMLElementForToken(t)
			c.activeFormattingElements.Push(elem)
		case "applet", "marquee", "object":
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			c.activeFormattingElements.Push(spec.ScopeMarker)
			c.frameset = framesetNotOK
		case "table":
			if c.quirksMode != quirks && c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inTable, noError
		case "area", "br", "embed", "img", "keygen", "wbr":
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			//ack token?
			c.frameset = framesetNotOK
		case "input":
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// ack self closing
			hasType := false
			for key, value := range t.Attributes {
				if key == "type" && !strings.EqualFold("hidden", value) {
					c.frameset = framesetNotOK
					hasType = true
				}
			}
			if !hasType {
				c.frameset = framesetNotOK
			}
		case "param", "source", "track":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// ack self closing
		case "hr":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// ack self closing flag
			c.frameset = framesetNotOK
		case "image":
			t.TagName = "img"
			return true, inBody, generalParseError
		case "textarea":
			c.insertHTMLElementForToken(t)
			c.switchTokenizerState(t, rcDataState)
			c.originalInsertionMode = c.curInsertionMode
			c.frameset = framesetNotOK
			return false, inBodyPeekNextToken, noError
		case "xmp":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.reconstructActiveFormattingElements()
			c.frameset = framesetNotOK
			return c.genericRawTextElementParsingAlgorithm(t)
		case "iframe":
			c.frameset = framesetNotOK
			return c.genericRawTextElementParsingAlgorithm(t)
		case "noembed":
			return c.genericRawTextElementParsingAlgorithm(t)
		case "noscript":
			if c.scriptingEnabled {
				return c.genericRawTextElementParsingAlgorithm(t)
			}
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
				c.stackOfOpenElements.Pop()
			}
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
		case "rb", "rtc":
		case "rp", "rt":
		case "math":
			c.reconstructActiveFormattingElements()
			c.adjustMathMLAttrs(t)
			// TODO: not really sure how namespaced attributes work here
			//c.adjustForeignAttributes(t)
			c.insertForeignElementForToken(t, spec.Mathmlns)
			if t.SelfClosing {
				c.stackOfOpenElements.Pop()
				//TODO: ack the self-closing tag
			}
		case "svg":
		case "caption", "col", "colgroup", "frame", "head", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inBody, generalParseError
		default:
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
		}
		return false, inBody, err
	case endTagToken:
		switch t.TagName {
		case "template":
		case "body":
			if !c.stackOfOpenElements.ContainsElementInScope("body") {
				return false, inBody, generalParseError
			}

			return false, afterBody, c.stackOfOpenElementsParseErrorCheck()
		case "html":
		case "address", "article", "aside", "blockquote", "button", "center", "details", "dialog",
			"dir", "div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup",
			"listing", "main", "menu", "nav", "ol", "pre", "section", "summary", "ul":

			if !c.stackOfOpenElements.ContainsElementInScope(webidl.DOMString(t.TagName)) {
				return false, inBody, generalParseError
			}

			c.generateImpliedEndTags([]webidl.DOMString{})
			if c.getCurrentNode().NodeName != webidl.DOMString(t.TagName) {
				err = generalParseError
			}
			c.stackOfOpenElements.PopUntil(webidl.DOMString(t.TagName))
		case "form":
			if len(c.containedInStackOpenElements("template")) == 0 {
				node := c.formElementPointer
				c.formElementPointer = nil
				if node == nil ||
					!c.stackOfOpenElements.ContainsElementInScope(node.NodeName) {
					return false, inBody, generalParseError
				}
				c.generateImpliedEndTags([]webidl.DOMString{})
				if c.getCurrentNode().NodeName != node.NodeName {
					err = generalParseError
				}
				c.stackOfOpenElements.Remove(c.stackOfOpenElements.Contains(node))
				return false, inBody, err
			}

			if !c.stackOfOpenElements.ContainsElementInScope("form") {
				return false, inBody, generalParseError
			}
			c.generateImpliedEndTags([]webidl.DOMString{})
			if c.getCurrentNode().NodeName != "form" {
				err = generalParseError
			}
			c.stackOfOpenElements.PopUntil("form")
			return false, inBody, err
		case "p":
			if !c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				err = generalParseError
				tok := &Token{
					TagName:   "p",
					TokenType: startTagToken,
				}
				c.insertHTMLElementForToken(tok)
			}

			c.closePElement()
		case "li":
			if !c.stackOfOpenElements.ContainsElementInListItemScope("li") {
				return false, inBody, generalParseError
			}
			c.generateImpliedEndTags([]webidl.DOMString{"li"})
			if c.getCurrentNode().NodeName != "li" {
				err = generalParseError
			}
			c.stackOfOpenElements.PopUntil("li")
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
			if !c.stackOfOpenElements.ContainsElementInScope(webidl.DOMString(t.TagName)) {
				return false, inBody, generalParseError
			}

			c.generateImpliedEndTags([]webidl.DOMString{})
			if c.getCurrentNode().NodeName == webidl.DOMString(t.TagName) {
				err = generalParseError
			}
			c.stackOfOpenElements.PopUntil(webidl.DOMString(t.TagName))
			c.clearListOfActiveFormattingElementsToLastMarker()
		case "br":
			t.TokenType = startTagToken
			t.Attributes = map[string]string{}
			return true, inBody, generalParseError
		default:
			err = c.defaultInBodyModeHandler(t)
		}

		return false, inBody, err
	case endOfFileToken:
		if len(c.stackOfTemplateInsertionModes) != 0 {
			return c.useRulesFor(t, inTemplate)
		}

		err := c.stackOfOpenElementsParseErrorCheck()
		return c.stopParsing(err)
	}
	return false, inBody, err
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
		c.stackOfOpenElements.Pop()
		return true, c.originalInsertionMode, generalParseError
	case endTagToken:
		switch t.TagName {
		case "script":
			c.stackOfOpenElements.Pop()
			return false, c.originalInsertionMode, noError
		default:
			c.stackOfOpenElements.Pop()
			return false, c.originalInsertionMode, noError
		}
	}
	return false, text, noError
}

func (c *HTMLTreeConstructor) defaultInTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	c.fosterParenting = true
	repro, nextState, _ := c.useRulesFor(t, inBody)
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
		return false, inTable, noError
	case docTypeToken:
		return false, inTable, generalParseError
	case startTagToken:
		switch t.TagName {
		case "caption":
			c.clearStackBackToTable()
			c.activeFormattingElements.Push(spec.ScopeMarker)
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
			if c.stackOfOpenElements.ContainsElementInTableScope("table") {
				c.stackOfOpenElements.PopUntil("table")
				mode = c.resetInsertionMode()
				repro = true
			}
			return repro, mode, noError
		case "style", "script", "template":
			return c.useRulesFor(t, inHead)
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
			c.stackOfOpenElements.Pop()
			// ack self closing
			return false, inTable, generalParseError
		case "form":
			if len(c.containedInStackOpenElements("template")) != 0 ||
				c.formElementPointer != nil {
				return false, inTable, generalParseError
			}

			elem := c.insertHTMLElementForToken(t)
			c.formElementPointer = elem
			c.stackOfOpenElements.Pop()
		}
	case endTagToken:
		switch t.TagName {
		case "table":
			mode := inTable
			if !c.stackOfOpenElements.ContainsElementInTableScope("table") {
				return false, inTable, generalParseError
			}

			c.stackOfOpenElements.PopUntil("table")
			mode = c.resetInsertionMode()
			return false, mode, noError
		case "body", "caption", "col", "colgroup", "html", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inTable, generalParseError
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inBody)
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

	for _, tok := range c.pendingTableCharacterTokens {
		if !isASCIIWhitespace(int(tok.Data[0])) {
			for _, t := range c.pendingTableCharacterTokens {
				c.defaultInTableModeHandler(t)
			}

			return true, c.originalInsertionMode, generalParseError
		}
	}

	for _, tok := range c.pendingTableCharacterTokens {
		c.insertCharacter(tok)
	}
	return true, c.originalInsertionMode, noError
}

func (c *HTMLTreeConstructor) inCaptionHelper() parseError {
	if !c.stackOfOpenElements.ContainsElementInTableScope("caption") {
		return generalParseError
	}

	c.generateImpliedEndTags([]webidl.DOMString{})
	err := noError
	if c.getCurrentNode().NodeName != "caption" {
		err = generalParseError
	}
	c.stackOfOpenElements.PopUntil("caption")
	c.clearListOfActiveFormattingElementsToLastMarker()
	return err
}

func (c *HTMLTreeConstructor) inCaptionModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "caption", "col", "colgroup", "tbody", "td", "tfoot", "th", "thead", "tr":
			return true, inTable, c.inCaptionHelper()
		}
	case endTagToken:
		switch t.TagName {
		case "caption":
			return false, inTable, c.inCaptionHelper()
		case "table":
			return true, inTable, c.inCaptionHelper()
		case "body", "col", "colgroup", "html", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inCaption, generalParseError
		}
	}
	return c.useRulesFor(t, inBody)
}
func (c *HTMLTreeConstructor) inColumnGroupModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			c.insertComment(t)
			return false, inColumnGroup, noError
		}
	case commentToken:
		c.insertComment(t)
		return false, inColumnGroup, noError
	case docTypeToken:
		return false, inColumnGroup, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "col":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// TODO: ack
			return false, inColumnGroup, noError
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endTagToken:
		switch t.TagName {
		case "colgroup":
			if c.getCurrentNode().NodeName == "colgroup" {
				return false, inColumnGroup, generalParseError
			}
			c.stackOfOpenElements.Pop()
			return false, inTable, noError
		case "col":
			return false, inColumnGroup, generalParseError
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inBody)
	}

	if c.getCurrentNode().NodeName != "colgroup" {
		return false, inColumnGroup, generalParseError
	}
	c.stackOfOpenElements.Pop()
	return true, inTable, noError
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
			if !c.stackOfOpenElements.ContainsElementInTableScope("tbody") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("thead") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("tfoot") {
				return false, inTableBody, generalParseError
			}

			c.clearStackBackToTableBody()
			c.stackOfOpenElements.Pop()
			return true, inTable, err
		}
	case endTagToken:
		switch t.TagName {
		case "tbody", "tfoot", "thead":
			if c.stackOfOpenElements.ContainsElementInTableScope(webidl.DOMString(t.TagName)) {
				c.clearStackBackToTableBody()
				c.stackOfOpenElements.Pop()
				return true, inTable, noError
			}

			return false, inTableBody, generalParseError
		case "table":
			if !c.stackOfOpenElements.ContainsElementInTableScope("tbody") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("thead") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("tfoot") {
				return false, inTableBody, generalParseError
			}
			c.clearStackBackToTableBody()
			c.stackOfOpenElements.Pop()
			return true, inTable, err
		case "body", "caption", "col", "colgroup", "html", "td", "th", "tr":
			return false, inTableBody, generalParseError
		}
	}
	return c.useRulesFor(t, inTable)
}
func (c *HTMLTreeConstructor) inRowModeHandler(t *Token) (bool, insertionMode, parseError) {
	err := noError
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "th", "td":
			c.clearStackBackToTableRow()
			c.insertHTMLElementForToken(t)
			c.activeFormattingElements.Push(spec.ScopeMarker)
			return false, inCell, err
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr":
			if c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				c.clearStackBackToTableRow()
				c.stackOfOpenElements.Pop()
				return true, inTableBody, noError
			}
			return false, inRow, generalParseError
		}
	case endTagToken:
		switch t.TagName {
		case "tr":
			if c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				c.clearStackBackToTableRow()
				c.stackOfOpenElements.Pop()
				return false, inTableBody, noError
			}

			return false, inRow, generalParseError
		case "table":
			if !c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				return false, inRow, generalParseError
			}
			c.clearStackBackToTableRow()
			c.stackOfOpenElements.Pop()
			return true, inTableBody, noError
		case "tbody", "tfoot", "thead":
			if !c.stackOfOpenElements.ContainsElementInTableScope(webidl.DOMString(t.TagName)) {
				return false, inRow, generalParseError
			}
			if !c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				return false, inRow, noError
			}

			c.clearStackBackToTableRow()
			c.stackOfOpenElements.Pop()
			return true, inTableBody, noError
		case "body", "caption", "col", "colgroup", "html", "td", "th":
			return false, inRow, generalParseError
		}
	}
	return c.useRulesFor(t, inTable)
}

func (c *HTMLTreeConstructor) closeCell() (bool, insertionMode, parseError) {
	err := noError
	c.generateImpliedEndTags([]webidl.DOMString{})
	cur := c.getCurrentNode().NodeName
	if cur != "td" && cur != "th" {
		err = generalParseError
	}
	c.stackOfOpenElements.PopUntilMany([]webidl.DOMString{"td", "th"})
	c.clearListOfActiveFormattingElementsToLastMarker()
	return true, inRow, err
}

func (c *HTMLTreeConstructor) inCellModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "caption", "col", "colgroup", "tbody", "td", "tfoot", "th", "thead", "tr":
			if !c.stackOfOpenElements.ContainsElementInTableScope("th") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("td") {
				return false, inCell, generalParseError
			}

			return c.closeCell()
		}
	case endTagToken:
		switch t.TagName {
		case "td", "th":
			if !c.stackOfOpenElements.ContainsElementInTableScope(webidl.DOMString(t.TagName)) {
				return false, inCell, generalParseError
			}
			err := noError
			c.generateImpliedEndTags([]webidl.DOMString{})
			if c.getCurrentNode().NodeName != webidl.DOMString(t.TagName) {
				err = generalParseError
			}
			c.stackOfOpenElements.PopUntil(webidl.DOMString(t.TagName))
			c.activeFormattingElements.PopUntil(spec.ScopeMarker.NodeName)
			return false, inRow, err
		case "body", "caption", "col", "colgroup", "html":
			return false, inCell, generalParseError
		case "table", "tbody", "tfoot", "thead", "tr":
			if !c.stackOfOpenElements.ContainsElementInTableScope(webidl.DOMString(t.TagName)) {
				return false, inCell, noError
			}
			return c.closeCell()
		}
	}
	return c.useRulesFor(t, inBody)
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
				c.stackOfOpenElements.Pop()
			}
			c.insertHTMLElementForToken(t)
		case "optgroup":
			if c.getCurrentNode().NodeName == "option" {
				c.stackOfOpenElements.Pop()
			}

			if c.getCurrentNode().NodeName == "optgroup" {
				c.stackOfOpenElements.Pop()
			}
			c.insertHTMLElementForToken(t)
		case "select":
			if !c.stackOfOpenElements.ContainsElementInSelectScope("select") {
				return false, inSelect, generalParseError
			}

			c.stackOfOpenElements.PopUntil("select")
			return false, c.resetInsertionMode(), generalParseError
		case "input", "keygen", "textarea":
			if !c.stackOfOpenElements.ContainsElementInSelectScope("select") {
				return false, inSelect, generalParseError
			}

			c.stackOfOpenElements.PopUntil("select")
			return true, c.resetInsertionMode(), generalParseError
		case "script", "template":
			return c.useRulesFor(t, inHead)
		}
	case endTagToken:
		switch t.TagName {
		case "optgroup":
			if c.getCurrentNode().NodeName == "option" &&
				len(c.stackOfOpenElements) >= 2 &&
				c.stackOfOpenElements[len(c.stackOfOpenElements)-2].NodeName == "optgroup" {
				c.stackOfOpenElements.Pop()
			}

			if c.getCurrentNode().NodeName != "optgroup" {
				return false, inSelect, generalParseError
			}
			c.stackOfOpenElements.Pop()
			return false, inSelect, noError
		case "option":
			if c.getCurrentNode().NodeName == "option" {
				c.stackOfOpenElements.Pop()
				return false, inSelect, noError
			}
			return false, inSelect, generalParseError
		case "select":
			if !c.stackOfOpenElements.ContainsElementInSelectScope("select") {
				return false, inSelect, generalParseError
			}

			c.stackOfOpenElements.PopUntil("select")
			return false, c.resetInsertionMode(), noError
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inBody)
	default:
		err = generalParseError
	}
	return false, inSelect, err
}
func (c *HTMLTreeConstructor) inSelectInTableModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case startTagToken:
		c.stackOfOpenElements.PopUntil("select")
		return true, c.resetInsertionMode(), generalParseError
	case endTagToken:
		if !c.stackOfOpenElements.ContainsElementInTableScope(webidl.DOMString(t.TagName)) {
			return false, inSelectInTable, generalParseError
		}
		c.stackOfOpenElements.PopUntil("select")
		return true, c.resetInsertionMode(), generalParseError
	}
	return c.useRulesFor(t, inSelect)
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
		c.stackOfOpenElements.PopUntil("template")
		c.clearListOfActiveFormattingElementsToLastMarker()
		c.stackOfTemplateInsertionModes = c.stackOfTemplateInsertionModes[:len(c.stackOfTemplateInsertionModes)-1]
		return true, c.resetInsertionMode(), generalParseError
	default:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) afterBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			return c.useRulesFor(t, inBody)
		}
	case commentToken:
		children := c.stackOfOpenElements[0].ChildNodes
		il := &insertionLocation{
			node: children[len(children)-1],
			insert: func(n *spec.Node) {
				c.stackOfOpenElements[0].AppendChild(n)
			},
		}
		c.insertCommentAt(t, il)
		return false, afterBody, noError
	case docTypeToken:
		return false, afterBody, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		}
	case endTagToken:
		if t.TagName == "html" {
			if c.createdBy == htmlFragmentParsingAlgorithm {
				return false, afterBody, generalParseError
			}
			return false, afterAfterBody, noError
		}
	case endOfFileToken:
		c.stopParsing(noError)
	}
	return true, inBody, generalParseError
}
func (c *HTMLTreeConstructor) inFramesetModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			c.insertCharacter(t)
		}
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		return false, inFrameset, generalParseError
	case startTagToken:
		switch t.TagName {
		case "frameset":
			c.insertHTMLElementForToken(t)
		case "frame":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// TODO: ack self-closing
		case "noframes":
			return c.useRulesFor(t, inHead)
		}
	case endTagToken:
		switch t.TagName {
		case "frameset":
			if c.getCurrentNode().NodeName == "html" {
				return false, inFrameset, generalParseError
			}
			c.stackOfOpenElements.Pop()
			if c.createdBy != htmlFragmentParsingAlgorithm &&
				c.getCurrentNode().NodeName != "frameset" {
				return false, afterFrameset, noError
			}
		}
	}
	return false, inFrameset, generalParseError
}
func (c *HTMLTreeConstructor) afterFramesetModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			c.insertCharacter(t)
		}
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		return false, afterFrameset, generalParseError
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "noframes":
			return c.useRulesFor(t, inHead)
		}
	case endTagToken:
		switch t.TagName {
		case "html":
			return false, afterAfterFrameset, noError
		}
	}
	return false, afterFrameset, generalParseError
}
func (c *HTMLTreeConstructor) afterAfterBodyModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			return c.useRulesFor(t, inBody)
		}
	case commentToken:
		il := &insertionLocation{
			node: c.HTMLDocument.Node,
			insert: func(n *spec.Node) {
				c.HTMLDocument.Node.AppendChild(n)
			},
		}
		c.insertCommentAt(t, il)
	case docTypeToken:
		return c.useRulesFor(t, inBody)
	case endOfFileToken:
		return c.stopParsing(noError)
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		}
	}
	return false, inBody, generalParseError
}
func (c *HTMLTreeConstructor) afterAfterFramesetModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			return c.useRulesFor(t, inBody)
		}
	case commentToken:
		il := &insertionLocation{
			node: c.HTMLDocument.Node,
			insert: func(n *spec.Node) {
				c.HTMLDocument.Node.AppendChild(n)
			},
		}
		c.insertCommentAt(t, il)
	case docTypeToken:
		return c.useRulesFor(t, inBody)
	case endOfFileToken:
		return c.stopParsing(noError)
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "noframes":
			return c.useRulesFor(t, inHead)
		}
	}
	return false, afterAfterFrameset, generalParseError
}

func (c *HTMLTreeConstructor) switchTokenizerState(t *Token, state tokenizerState) {
	c.stateChannel <- state
	t.Special = false
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
	inBodyPeekNextToken
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
	c.constructTreeStartState(initial)
}

func (c *HTMLTreeConstructor) constructTreeStartState(im insertionMode) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s %s\n", err, de.Stack())
		}
		c.wg.Done()
	}()

	var (
		token    *Token
		nextMode insertionMode = im
	)
	c.curInsertionMode = im
	for token = range c.tokenChannel {
		reprocess := true
		if nextMode == stopParser {
			break
		}
		for reprocess {
			reprocess, nextMode = c.processToken(token, nextMode)
			c.curInsertionMode = nextMode
		}
	}
}

var tagStateMappings = map[string]insertionMode{
	"script":    inHead,
	"noembed":   inBody,
	"textarea":  inBody,
	"iframe":    inBody,
	"noframes":  inHead,
	"style":     inHead,
	"title":     inHead,
	"plaintext": inBody,
	"xmp":       inBody,
}

// A special token is a start tag token in either the inHead or inBody mode
// and in the list of token keys above.
func isSpecialToken(token *Token) bool {
	return token.Special
}

func isWrongTreeState(token *Token, nextMode insertionMode, scriptingEnabled bool) bool {
	// quick mode check before other odd state checks
	// if we aren't in the head or body mode handlers, we
	// are in the wrong state for a special token.
	if nextMode != inHead && nextMode != inBody {
		return true
	}

	// special check for noscript elements because we need to be able to tell if
	// scripting has been enabled or not for these elements. Also there are two
	// states that we need to check for noscript elements
	if token.TagName == "noscript" {
		if !scriptingEnabled {
			return true
		} else {
			return false
		}
	}

	return nextMode != tagStateMappings[token.TagName]
}

func specialTokenWrongState(token *Token, nextMode insertionMode, scriptingEnabled bool) bool {
	return isSpecialToken(token) && isWrongTreeState(token, nextMode, scriptingEnabled)
}

func isMathmlIntPoint(e *spec.Node) bool {
	switch e.TagName {
	case "mi", "mo", "mn", "ms", "mtext":
		return true
	}
	return false
}

func isHTMLIntPoint(e *spec.Node) bool {
	if e.TagName == "annotation-xml" && e.Element.NamespaceURI == spec.Mathmlns {
		if val, ok := e.Attributes.Attrs["encoding"]; ok {
			if val == "text/html" || val == "application/xhtml+xml" {
				return true
			}
		}
	}

	if e.Element.NamespaceURI == spec.Svgns {
		if e.TagName == "foreignObject" {
			return true
		}

		if e.TagName == "desc" {
			return true
		}

		if e.TagName == "title" {
			return true
		}
	}
	return false
}

var svgTable = map[string]string{
	"fedistantlight":     "feDistantLight",
	"fedropshadow":       "feDropShadow",
	"feflood":            "feFlood",
	"fefunca":            "feFuncA",
	"fefuncb":            "feFuncB",
	"fefuncg":            "feFuncG",
	"fefuncr":            "feFuncR",
	"fegaussianblur":     "feGaussianBlur",
	"feimage":            "feImage",
	"femerge":            "feMerge",
	"femergenode":        "feMergeNode",
	"femorphology":       "feMorphology",
	"feoffset":           "feOffset",
	"fepointlight":       "fePointLight",
	"fespecularlighting": "feSpecularLighting",
	"fespotlight":        "feSpotLight",
	"fetile":             "feTile",
	"feturbulence":       "feTurbulence",
	"foreignobject":      "foreignObject",
	"glyphref":           "glyphRef",
	"lineargradient":     "linearGradient",
	"radialgradient":     "radialGradient",
	"textpath":           "textPath",
}

func (c *HTMLTreeConstructor) defaultParseTokensInForeignContentEndScriptTag(t *Token, startMode insertionMode) (bool, insertionMode, parseError) {
	c.stackOfOpenElements.Pop()
	// insertion point
	// parser pause flag
	// process svg script tags
	return false, startMode, noError
}

func (c *HTMLTreeConstructor) defaultParseTokensInForeignContentStartTag(t *Token, startMode insertionMode) (bool, insertionMode, parseError) {
	switch c.getAdjustedCurrentNode().Element.NamespaceURI {
	case spec.Mathmlns:
		c.adjustMathMLAttrs(t)
	case spec.Svgns:
		if val, ok := svgTable[t.TagName]; ok {
			t.TagName = val
		}

	}

	c.adjustSVGAttrs(t)
	//TODO: c.adjustForeignAttributes
	c.insertForeignElementForToken(t, c.getAdjustedCurrentNode().Element.NamespaceURI)
	if t.SelfClosing {
		// todo: acl self closing
		return c.defaultParseTokensInForeignContentEndScriptTag(t, startMode)
	}
	c.stackOfOpenElements.Pop()
	//todo: ack self closing
	return false, startMode, noError
}

func (c *HTMLTreeConstructor) parseTokensInForeignContent(t *Token, startMode insertionMode) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0000":
			t.Data = "\uFFFD"
			c.insertCharacter(t)
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.insertCharacter(t)
		default:
			c.insertCharacter(t)
			c.frameset = framesetNotOK
		}

		return false, startMode, noError
	case commentToken:
		c.insertComment(t)
		return false, startMode, noError
	case docTypeToken:
		return false, startMode, generalParseError
	case startTagToken:
		switch t.TagName {
		case "b", "big", "blockquote", "body", "br", "center", "code", "dd", "div",
			"dl", "dt", "em", "embed", "h1", "h2", "h3", "h4", "h5", "h6", "head",
			"hr", "i", "img", "li", "listing", "menu", "meta", "nobr", "ol", "p",
			"pre", "ruby", "s", "small", "span", "strong", "strike", "sub", "sup",
			"table", "tt", "u", "ul", "var":
			if c.createdBy == htmlFragmentParsingAlgorithm {
				return c.defaultParseTokensInForeignContentStartTag(t, startMode)
			}
			c.stackOfOpenElements.Pop()
			c.stackOfOpenElements.PopUntilConditions([]func(e *spec.Node) bool{
				isHTMLIntPoint,
				isMathmlIntPoint,
				func(e *spec.Node) bool { return e.Element != nil && e.Element.NamespaceURI == spec.Htmlns },
			})
			return true, startMode, generalParseError
		case "font":
			for k := range t.Attributes {
				switch k {
				case "color", "face", "size":
					if c.createdBy == htmlFragmentParsingAlgorithm {
						return c.defaultParseTokensInForeignContentStartTag(t, startMode)
					}
					c.stackOfOpenElements.Pop()
					c.stackOfOpenElements.PopUntilConditions([]func(e *spec.Node) bool{
						isHTMLIntPoint,
						isMathmlIntPoint,
						func(e *spec.Node) bool { return e.Element != nil && e.Element.NamespaceURI == spec.Htmlns },
					})
					return true, startMode, generalParseError
				}
			}
		default:
			return c.defaultParseTokensInForeignContentStartTag(t, startMode)
		}
	case endTagToken:
		if t.TagName == "script" && c.getCurrentNode().NodeName == "svg" && c.getCurrentNode().Element.NamespaceURI == spec.Svgns {
			return c.defaultParseTokensInForeignContentEndScriptTag(t, startMode)
		}

		err := noError
		last := len(c.stackOfOpenElements) - 1
		for i := last; i >= 1; i-- {
			node := c.stackOfOpenElements[i]
			if i != last && node.Element.NamespaceURI == spec.Htmlns {
				return true, startMode, err
			}
			if strings.EqualFold(string(node.TagName), t.TagName) {
				for {
					popped := c.stackOfOpenElements.Pop()
					if popped == node {
						return false, startMode, err
					}
				}
			} else {
				if i == last {
					err = generalParseError
				}
			}
		}
	}

	return false, startMode, noError
}

func (c *HTMLTreeConstructor) dispatch(t *Token, startMode insertionMode) (bool, insertionMode, parseError) {
	acn := c.getAdjustedCurrentNode()
	if len(c.stackOfOpenElements) == 0 ||
		acn.Element.NamespaceURI == spec.Htmlns ||
		(t.TagName != "mglyph" && t.TagName != "malignmark" && t.TokenType == startTagToken && isMathmlIntPoint(acn)) ||
		(isMathmlIntPoint(acn) && t.TokenType == characterToken) ||
		(acn.TagName == "annotation-xml" && t.TokenType == startTagToken && t.TagName == "svg") ||
		(isHTMLIntPoint(acn) && (t.TokenType == startTagToken || t.TokenType == characterToken)) {
		return c.mappings[startMode](t)
	}

	return parseTokensInForeignContent(t)
}

func (c *HTMLTreeConstructor) processToken(t *Token, startMode insertionMode) (bool, insertionMode) {
	fmt.Printf("[TREE]token: %+vmode: %s\n", t, startMode)
	reprocess, nextMode, parseErr := c.dispatch(t, startMode)
	fmt.Printf("[TREE]tree after: \n%s\n\n", c.HTMLDocument.Node)
	if c.config[debug] == 0 {
		logError(parseErr)
	}
	// if we didn't consume the token, we don't want to check this state
	// only check if the token is in this special state when we are in our
	// consuming token state
	if !reprocess && specialTokenWrongState(t, startMode, c.scriptingEnabled) {
		c.switchTokenizerState(t, dataState)
	}
	return reprocess, nextMode
}
