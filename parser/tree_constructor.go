package parser

import (
	"fmt"
	"strings"

	"github.com/heathj/gobrowse/parser/spec"
)

type quirksMode uint

const (
	noQuirks quirksMode = iota
	quirks
	limitedQuirks
)

type frameset uint

const (
	framesetOK frameset = iota
	framesetNotOK
)

// HTMLTreeConstructor holds the state for various state of the tree construction phase.
type HTMLTreeConstructor struct {
	nextTokenizerState                              *tokenizerState
	curInsertionMode                                insertionMode
	HTMLDocument                                    *spec.HTMLDocument
	quirksMode                                      quirksMode
	fosterParenting, scriptingEnabled               bool
	originalInsertionMode                           insertionMode
	stackOfOpenElements, activeFormattingElements   spec.NodeList
	stackOfTemplateInsertionModes                   []insertionMode
	headElementPointer, formElementPointer, context *spec.Node
	pendingTableCharacterTokens                     []Token
	frameset                                        frameset
}

// NewHTMLTreeConstructor creates an HTMLTreeConstructor.
func NewHTMLTreeConstructor() *HTMLTreeConstructor {
	return &HTMLTreeConstructor{
		HTMLDocument: spec.NewHTMLDocumentNode(),
	}
}

func (c *HTMLTreeConstructor) modeToModeHandler(mode insertionMode) treeConstructionModeHandler {
	switch mode {
	case initial:
		return c.initialModeHandler
	case beforeHTML:
		return c.beforeHTMLModeHandler
	case beforeHead:
		return c.beforeHeadModeHandler
	case inHead:
		return c.inHeadModeHandler
	case inHeadNoScript:
		return c.inHeadNoScriptModeHandler
	case afterHead:
		return c.afterHeadModeHandler
	case inBody:
		return c.inBodyModeHandler
	case inBodyPeekNextToken:
		return c.inBodyPeekNextToken
	case text:
		return c.textModeHandler
	case inTable:
		return c.inTableModeHandler
	case inTableText:
		return c.inTableTextModeHandler
	case inCaption:
		return c.inCaptionModeHandler
	case inColumnGroup:
		return c.inColumnGroupModeHandler
	case inTableBody:
		return c.inTableBodyModeHandler
	case inRow:
		return c.inRowModeHandler
	case inCell:
		return c.inCellModeHandler
	case inSelect:
		return c.inSelectModeHandler
	case inSelectInTable:
		return c.inSelectInTableModeHandler
	case inTemplate:
		return c.inTemplateModeHandler
	case afterBody:
		return c.afterBodyModeHandler
	case inFrameset:
		return c.inFramesetModeHandler
	case afterFrameset:
		return c.afterFramesetModeHandler
	case afterAfterBody:
		return c.afterAfterBodyModeHandler
	case afterAfterFrameset:
		return c.afterAfterFramesetModeHandler
	}
	return nil
}

func (c *HTMLTreeConstructor) getCurrentNode() *spec.Node {
	if len(c.stackOfOpenElements) == 0 {
		return nil
	}
	return c.stackOfOpenElements[len(c.stackOfOpenElements)-1]
}

func (c *HTMLTreeConstructor) getAdjustedCurrentNode() *spec.Node {
	if c.context == nil {
		return c.getCurrentNode()
	}

	return c.context
}

// Inserts a comment at a specific location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertCommentAt(t Token, il *insertionLocation) {
	commentNode := spec.NewComment(t.Data, il.node.OwnerDocument)
	il.insert(commentNode)
}

// Inserts a comment at the adjusted insertion location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertComment(t Token) {
	c.insertCommentAt(t, c.getAppropriatePlaceForInsertion(nil))
}

func (c *HTMLTreeConstructor) getLastElemInStackOfOpenElements(elem string) int {
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

type foreignAttrTableEntry struct {
	prefix, localName string
	ns                spec.Namespace
}

var foreignAttrsTable = map[string]foreignAttrTableEntry{
	"xlink:actuate": {"xlink", "actuate", spec.Xlinkns},
	"xlink:arcrole": {"xlink", "arcrole", spec.Xlinkns},
	"xlink:href":    {"xlink", "href", spec.Xlinkns},
	"xlink:role":    {"xlink", "role", spec.Xlinkns},
	"xlink:show":    {"xlink", "show", spec.Xlinkns},
	"xlink:title":   {"xlink", "title", spec.Xlinkns},
	"xlink:type":    {"xlink", "type", spec.Xlinkns},
	"xml:lang":      {"xml", "lang", spec.Xmlns},
	"xml:space":     {"xml", "space", spec.Xmlns},
	"xmlns":         {"", "actuate", spec.Xmlnsns},
	"xmlns:xlink":   {"xmlns", "xlink", spec.Xmlnsns},
}

func (c *HTMLTreeConstructor) adjustForeignAttributes(t Token) {
	for k, attr := range t.Attributes {
		if entry, ok := foreignAttrsTable[k]; ok {
			attr.LocalName = entry.localName
			attr.Prefix = entry.prefix
			attr.Namespace = entry.ns
			t.Attributes[string(attr.LocalName)] = attr
			delete(t.Attributes, k)
		}
	}
}

func (c *HTMLTreeConstructor) adjustMathMLAttrs(t Token) {
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
	"diffuseconstant":     "diffuseConstant",
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

func (c *HTMLTreeConstructor) adjustSVGAttrs(t Token) {
	for k := range t.Attributes {
		if val, ok := svgAttrTable[k]; ok {
			t.Attributes[val] = t.Attributes[k]
			t.Attributes[val].LocalName = val
			delete(t.Attributes, k)
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
	}
	ail.node = target
	ail.insert = func(n *spec.Node) { target.AppendChild(n) }

	return ail
}

// https://html.spec.whatwg.org/multipage/custom-elements.html#custom-element-definition
type CustomElementDefinition struct {
}

//https://html.spec.whatwg.org/multipage/custom-elements.html#look-up-a-custom-element-definition
func (c *HTMLTreeConstructor) lookUpCustomElementDefinition(document *spec.Node, ns spec.Namespace, localName, is string) *CustomElementDefinition {
	//TODO:
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
			if context != nil {
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
func (c *HTMLTreeConstructor) stopParsing() (bool, insertionMode) {
	for len(c.stackOfOpenElements) > 0 {
		c.stackOfOpenElements.Pop()
	}

	return false, stopParser
}

// createElementForToken creates an element from a token with the provided
// namespace and parent element.
// https://html.spec.whatwg.org/multipage/parsing.html#create-an-element-for-the-token
func (c *HTMLTreeConstructor) createElementForToken(t Token, ns spec.Namespace, ip *spec.Node) *spec.Node {
	document := ip.OwnerDocument
	localName := t.TagName
	is, ok := t.Attributes["is"]
	var definition *CustomElementDefinition
	executeScript := false
	if ok {
		// won't need to implement this for a while
		definition = c.lookUpCustomElementDefinition(document, ns, localName, is.Value)
	}
	if definition != nil && c.context != nil {
		executeScript = true
	}

	if executeScript {
		//TODO: executeScript
	}

	element := spec.NewDOMElement(document, localName, ns)
	element.Attributes = spec.NewNamedNodeMap(t.Attributes, element)
	element.ParentNode = ip.ParentNode
	return element
}

func (c *HTMLTreeConstructor) insertCharacter(t Token) {
	il := c.getAppropriatePlaceForInsertion(nil)
	if il.node != nil && il.node.NodeType == spec.DocumentNode {
		return
	}

	tn := spec.NewTextNode(il.node.OwnerDocument, t.Data)
	il.insert(tn)
	if tn.PreviousSibling != nil && tn.PreviousSibling.NodeType == spec.TextNode {
		tn.PreviousSibling.Text.CharacterData.Data += t.Data
		il.node.RemoveChild(tn)
	} else if tn.NextSibling != nil && tn.NextSibling.NodeType == spec.TextNode {
		tn.NextSibling.Text.CharacterData.Data = t.Data + tn.NextSibling.Text.CharacterData.Data
		il.node.RemoveChild(tn)
	}
}

func (c *HTMLTreeConstructor) insertHTMLElementForToken(t Token) *spec.Node {
	return c.insertForeignElementForToken(t, spec.Htmlns)
}

func (c *HTMLTreeConstructor) insertForeignElementForToken(t Token, namespace spec.Namespace) *spec.Node {
	il := c.getAppropriatePlaceForInsertion(nil)
	elem := c.createElementForToken(t, namespace, il.node)
	il.insert(elem)
	c.stackOfOpenElements.Push(elem)
	return elem
}

func (c *HTMLTreeConstructor) useRulesFor(t Token, mode insertionMode) (bool, insertionMode) {
	reprocess, nextMode := c.processToken(t, mode)

	// if the mode didn't change return to where we came from.
	if nextMode == mode {
		return reprocess, c.curInsertionMode
	}
	return reprocess, nextMode
}

func compareLastN(n int, elems spec.NodeList, elem *spec.Node) bool {
	last := len(elems) - 1
	lastElem := elems[last]
	for i := last - 1; i >= last-n; i-- {
		if elems[i].NodeName != lastElem.NodeName {
			return false
		}

		if elems[i].Element.NamespaceURI != lastElem.Element.NamespaceURI {
			return false
		}

		if elems[i].Attributes.Length != lastElem.Attributes.Length {
			return false
		}

		for k, v := range lastElem.Attributes.Attrs {
			e := elems[i].Attributes.GetNamedItem(k)
			if e == nil {
				return false
			}
			if v.Namespace != e.Namespace {
				return false
			}

			if v.Name != e.Name {
				return false
			}

			if v.Value != e.Value {
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

func (c *HTMLTreeConstructor) generateImpliedEndTags(denylist ...string) {
	for {
		nn := c.getCurrentNode().NodeName
		switch nn {
		case "dd", "dt", "li", "optgroup", "option", "p", "rb", "rp", "rt", "rtc":
			for _, b := range denylist {
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
	c.generateImpliedEndTags("p")
	// skipping parse error check
	c.stackOfOpenElements.PopUntil("p")
}

func (c *HTMLTreeConstructor) adoptionAgencyAlgorithm(t Token) bool {
	cur := c.getCurrentNode()
	if cur.NodeName == t.TagName && c.activeFormattingElements.Contains(cur) == -1 {
		c.stackOfOpenElements.Pop()
		return false
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

			if c.activeFormattingElements[y].NodeName == t.TagName {
				formattingElement = c.activeFormattingElements[y]
				break
			}
		}

		if formattingElement == nil {
			return true
		}

		// 7
		si = c.stackOfOpenElements.Contains(formattingElement)
		if si == -1 {
			// parse error
			c.activeFormattingElements.Remove(y)
			return false
		}

		// 8
		if !c.stackOfOpenElements.ContainsElementInScope(formattingElement.NodeName) {
			// parse error
			return false
		}

		// 9 just an error check

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
					return false
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

	return false
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
	// 5. Let entry be the entry one earlier than entry in the list of active formatting elements.
	for ; i >= 0; i-- {
		// 6. If entry is neither a marker nor an element that is also in the stack of open elements,
		// go to the step labeled rewind.
		doesContain = c.stackOfOpenElements.Contains(c.activeFormattingElements[i])
		if c.activeFormattingElements[i].NodeType == spec.ScopeMarkerNode || doesContain != -1 {
			break
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

func (c *HTMLTreeConstructor) isForceQuirks(t Token) bool {
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

func (c *HTMLTreeConstructor) isLimitedQuirks(t Token) bool {
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

func (c *HTMLTreeConstructor) defaultInitialModeHandler() (bool, insertionMode) {
	// TODO:only if not an iframe src document
	c.quirksMode = quirks
	return true, beforeHTML
}

// https://html.spec.whatwg.org/multipage/parsing.html#the-initial-insertion-mode
func (c *HTMLTreeConstructor) initialModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, initial
		}
	case commentToken:
		c.insertComment(t)
		return false, initial
	case docTypeToken:
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

		return false, beforeHTML
	}
	return c.defaultInitialModeHandler()
}

func (c *HTMLTreeConstructor) defaultBeforeHTMLModeHandler(t Token) (bool, insertionMode) {
	n := spec.NewDOMElement(c.HTMLDocument.Node, "html", spec.Htmlns)
	n.OwnerDocument = c.HTMLDocument.Node
	c.HTMLDocument.AppendChild(n)
	c.stackOfOpenElements.Push(n)

	// TODO: application cache algo
	return true, beforeHead
}

// https://html.spec.whatwg.org/multipage/parsing.html#the-before-html-insertion-mode
func (c *HTMLTreeConstructor) beforeHTMLModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case docTypeToken:
		return false, beforeHTML
	case commentToken:
		il := &insertionLocation{
			node: c.HTMLDocument.Node,
			insert: func(n *spec.Node) {
				c.HTMLDocument.Node.AppendChild(n)
			},
		}
		c.insertCommentAt(t, il)
		return false, beforeHTML
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, beforeHTML
		}
	case startTagToken:
		if t.TagName == "html" {
			elem := c.createElementForToken(t, spec.Htmlns, c.HTMLDocument.Node)
			c.HTMLDocument.AppendChild(elem)
			c.stackOfOpenElements.Push(elem)
			// handle navigation of a browsing context
			return false, beforeHead
		}
	case endTagToken:
		switch t.TagName {
		case "head", "body", "html", "br":
			return c.defaultBeforeHTMLModeHandler(t)
		default:
			return false, beforeHTML
		}
	}
	return c.defaultBeforeHTMLModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultBeforeHeadModeHandler(t Token) (bool, insertionMode) {
	elem := c.insertHTMLElementForToken(Token{
		TokenType: startTagToken,
		TagName:   "head",
	})
	c.headElementPointer = elem
	return true, inHead
}
func (c *HTMLTreeConstructor) beforeHeadModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, beforeHead
		}
	case commentToken:
		c.insertComment(t)
		return false, beforeHead
	case docTypeToken:
		return false, beforeHead
	case startTagToken:
		if t.TagName == "html" {
			return c.useRulesFor(t, inBody)
		}

		if t.TagName == "head" {
			elem := c.insertHTMLElementForToken(t)
			c.headElementPointer = elem
			return false, inHead
		}
	case endTagToken:
		switch t.TagName {
		case "head", "body", "html", "br":
			return c.defaultBeforeHeadModeHandler(t)
		}

		return false, beforeHead
	}

	return c.defaultBeforeHeadModeHandler(t)

}

func (c *HTMLTreeConstructor) genericRCDATAElementParsingAlgorithm(t Token) (bool, insertionMode) {
	c.insertHTMLElementForToken(t)
	c.originalInsertionMode = c.curInsertionMode
	c.switchTokenizerState(rcDataState)
	return false, text
}

func (c *HTMLTreeConstructor) genericRawTextElementParsingAlgorithm(t Token) (bool, insertionMode) {
	c.insertHTMLElementForToken(t)
	c.originalInsertionMode = c.curInsertionMode
	c.switchTokenizerState(rawTextState)
	return false, text
}

func (c *HTMLTreeConstructor) defaultInHeadModeHandler(t Token) (bool, insertionMode) {
	c.stackOfOpenElements.Pop()
	return true, afterHead
}
func (c *HTMLTreeConstructor) inHeadModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.insertCharacter(t)
			return false, inHead
		}
	case commentToken:
		c.insertComment(t)
		return false, inHead
	case docTypeToken:
		return false, inHead
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "base", "basefont", "bgsound", "link":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			return false, inHead
			//TODO: acknowledge the self closing flag?
		case "meta":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			//TODO: acknowledge the self closing flag?
			//TODO: char encoding settings
			return false, inHead
		case "title":
			return c.genericRCDATAElementParsingAlgorithm(t)
		case "noscript":
			if c.scriptingEnabled {
				return c.genericRawTextElementParsingAlgorithm(t)
			}
			c.insertHTMLElementForToken(t)
			return false, inHeadNoScript
		case "noframes", "style":
			return c.genericRawTextElementParsingAlgorithm(t)
		case "script":
			il := c.getAppropriatePlaceForInsertion(nil)
			elem := c.createElementForToken(t, spec.Htmlns, il.node)
			elem.ParserDocument = c.HTMLDocument
			elem.NonBlocking = false
			il.insert(elem)
			c.stackOfOpenElements.Push(elem)
			c.switchTokenizerState(scriptDataState)
			c.originalInsertionMode = c.curInsertionMode
			return false, text
		case "template":
		case "head":
			return false, inHead
		}
	case endTagToken:
		switch t.TagName {
		case "head":
			c.stackOfOpenElements.Pop()
			return false, afterHead
		case "body", "html", "br":
			return c.defaultInHeadModeHandler(t)
		case "template":
		default:
			return false, inHead
		}
	}

	return c.defaultInHeadModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultInHeadNoScriptModeHandler(t Token) (bool, insertionMode) {
	c.stackOfOpenElements.Pop()
	return true, inHead
}
func (c *HTMLTreeConstructor) inHeadNoScriptModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return c.useRulesFor(t, inHead)
		}
	case commentToken:
		return c.useRulesFor(t, inHead)
	case docTypeToken:
		return false, inHeadNoScript
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "basefont", "bgsound", "link", "meta", "noframe", "style":
			return c.useRulesFor(t, inHead)
		case "head", "noscript":
			return false, inHeadNoScript
		}
	case endTagToken:
		switch t.TagName {
		case "noscript":
			c.stackOfOpenElements.Pop()
			return false, inHead
		case "br":
			return c.defaultInHeadNoScriptModeHandler(t)
		default:
			return false, inHeadNoScript
		}
	}
	return c.defaultInHeadNoScriptModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultAfterHeadModeHandler(t Token) (bool, insertionMode) {
	c.insertHTMLElementForToken(Token{
		TokenType: startTagToken,
		TagName:   "body",
	})
	return true, inBody
}
func (c *HTMLTreeConstructor) afterHeadModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.insertCharacter(t)
			return false, afterHead
		}
	case commentToken:
		c.insertComment(t)
		return false, afterHead
	case docTypeToken:
		return false, afterHead
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "body":
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inBody
		case "frameset":
			c.insertHTMLElementForToken(t)
			return false, inFrameset
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style", "template", "title":
			c.stackOfOpenElements.Push(c.headElementPointer)
			repro, nextState := c.useRulesFor(t, inHead)
			c.stackOfOpenElements.Remove(c.stackOfOpenElements.Contains(c.headElementPointer))
			return repro, nextState
		case "head":
			return false, afterHead
		}
	case endTagToken:
		switch t.TagName {
		case "template":
			return c.useRulesFor(t, inHead)
		case "body", "html", "br":
			return c.defaultAfterHeadModeHandler(t)
		default:
			return false, afterHead
		}

	}
	return c.defaultAfterHeadModeHandler(t)
}

func (c *HTMLTreeConstructor) defaultInBodyModeHandler(t Token) {
	for i := len(c.stackOfOpenElements) - 1; i >= 0; i-- {
		node := c.stackOfOpenElements[i]
		if node.NodeName == t.TagName {
			c.generateImpliedEndTags(node.NodeName)

			for {
				popped := c.stackOfOpenElements.Pop()
				if popped == nil || popped == node {
					return
				}
			}
		} else {
			if isSpecial(node) {
				return
			}
		}
	}
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

func (c *HTMLTreeConstructor) inBodyPeekNextToken(t Token) (bool, insertionMode) {
	ret := inBody
	if c.getCurrentNode().NodeName == "textarea" {
		ret = text
	}
	if t.TokenType == characterToken && t.Data == "\u000A" {
		return false, ret
	}
	return true, ret
}

func (c *HTMLTreeConstructor) inBodyModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0000":
			return false, inBody
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
		return false, inBody
	case docTypeToken:
		return false, inBody
	case startTagToken:
		switch t.TagName {
		case "html":
			if len(c.containedInStackOpenElements("template")) > 0 {
				return false, inBody
			}

			for k, v := range t.Attributes {
				if attr := c.stackOfOpenElements[0].Attributes.GetNamedItem(k); attr == nil {
					c.stackOfOpenElements[0].Attributes.SetNamedItem(spec.NewAttr(k, v, c.stackOfOpenElements[0]))
				}
			}
			return false, inBody
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style",
			"template", "title":
			return c.useRulesFor(t, inHead)
		case "body":
			if len(c.stackOfOpenElements) <= 1 ||
				c.stackOfOpenElements[1].NodeName != "body" ||
				len(c.containedInStackOpenElements("template")) != 0 {
				return false, inBody
			}

			c.frameset = framesetNotOK
			for k, v := range t.Attributes {
				if attr := c.stackOfOpenElements[1].Attributes.GetNamedItem(k); attr == nil {
					c.stackOfOpenElements[1].Attributes.SetNamedItem(spec.NewAttr(k, v, nil))
				}
			}
		case "frameset":
			if len(c.stackOfOpenElements) <= 1 ||
				c.stackOfOpenElements[1].NodeName != "body" {
				return false, inBody
			}
			if c.frameset == framesetNotOK {
				return false, inBody
			}
			c.stackOfOpenElements[1].ParentNode.RemoveChild(c.stackOfOpenElements[1])
			for c.getCurrentNode().NodeName != "html" {
				c.stackOfOpenElements.Pop()
			}
			c.insertHTMLElementForToken(t)
			return false, inFrameset
		case "address", "article", "aside", "blockquote", "center", "details", "dialog", "dir",
			"div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup", "main",
			"menu", "nav", "ol", "p", "section", "summary", "ul":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
		case "h1", "h2", "h3", "h4", "h5", "h6":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			switch c.getCurrentNode().NodeName {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				c.stackOfOpenElements.Pop()
			}

			c.insertHTMLElementForToken(t)
			return false, inBody
		case "pre", "listing":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			c.frameset = framesetNotOK
			return false, inBodyPeekNextToken
		case "form":
			noTemp := len(c.containedInStackOpenElements("template")) == 0
			if c.formElementPointer != nil && noTemp {
				return false, inBody
			}

			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			elem := c.insertHTMLElementForToken(t)
			if noTemp {
				c.formElementPointer = elem
			}
		case "li":
			done := func() (bool, insertionMode) {
				if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
					c.closePElement()
				}
				c.insertHTMLElementForToken(t)
				return false, inBody
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

			c.generateImpliedEndTags("li")
			c.stackOfOpenElements.PopUntil("li")

			// done
			return done()
		case "dd", "dt":
			c.frameset = framesetNotOK
			var node *spec.Node
			for i := len(c.stackOfOpenElements) - 1; i >= 0; i-- {
				node = c.stackOfOpenElements[i]
				if node.NodeName == "dd" {
					c.generateImpliedEndTags("dd")
					c.stackOfOpenElements.PopUntil("dd")
					break
				}

				if node.NodeName == "dt" {
					c.generateImpliedEndTags("dt")
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
			return false, inBody
		case "plaintext":
			if c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.closePElement()
			}
			c.insertHTMLElementForToken(t)
			c.switchTokenizerState(plaintextState)
		case "button":
			if c.stackOfOpenElements.ContainsElementInScope("button") {
				c.generateImpliedEndTags()
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
					if c.adoptionAgencyAlgorithm(t) {
						c.defaultInBodyModeHandler(t)
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
			return false, inTable
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
			if attr, ok := t.Attributes["type"]; ok && !strings.EqualFold("hidden", string(attr.Value)) {
				c.frameset = framesetNotOK
				hasType = true
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
			return c.useRulesFor(t, inBody)
		case "textarea":
			c.insertHTMLElementForToken(t)
			c.switchTokenizerState(rcDataState)
			c.originalInsertionMode = c.curInsertionMode
			c.frameset = framesetNotOK
			return false, inBodyPeekNextToken
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
				return false, inSelectInTable
			}
			return false, inSelect
		case "optgroup", "option":
			if c.getCurrentNode().NodeName == "option" {
				c.stackOfOpenElements.Pop()
			}
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
		case "rb", "rtc":
			if c.stackOfOpenElements.ContainsElementInScope("ruby") {
				c.generateImpliedEndTags()
			}
			c.insertHTMLElementForToken(t)
		case "rp", "rt":
			if c.stackOfOpenElements.ContainsElementInScope("ruby") {
				c.generateImpliedEndTags("rtc")
			}
			c.insertHTMLElementForToken(t)
		case "math":
			c.reconstructActiveFormattingElements()
			c.adjustMathMLAttrs(t)
			c.adjustForeignAttributes(t)
			c.insertForeignElementForToken(t, spec.Mathmlns)
			if t.SelfClosing {
				c.stackOfOpenElements.Pop()
				//TODO: ack the self-closing tag
			}
			return false, inBody
		case "svg":
			c.reconstructActiveFormattingElements()
			c.adjustSVGAttrs(t)
			c.adjustForeignAttributes(t)
			c.insertForeignElementForToken(t, spec.Svgns)
			if t.SelfClosing {
				c.stackOfOpenElements.Pop()
				//TODO: ack the self-closing
			}
			return false, inBody
		case "caption", "col", "colgroup", "frame", "head", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inBody
		default:
			c.reconstructActiveFormattingElements()
			c.insertHTMLElementForToken(t)
		}
		return false, inBody
	case endTagToken:
		switch t.TagName {
		case "template":
			return c.useRulesFor(t, inHead)
		case "body":
			if !c.stackOfOpenElements.ContainsElementInScope("body") {
				return false, inBody
			}
			return false, afterBody
		case "html":
			if !c.stackOfOpenElements.ContainsElementInScope("body") {
				return false, inBody
			}

			return true, afterBody
		case "address", "article", "aside", "blockquote", "button", "center", "details", "dialog",
			"dir", "div", "dl", "fieldset", "figcaption", "figure", "footer", "header", "hgroup",
			"listing", "main", "menu", "nav", "ol", "pre", "section", "summary", "ul":

			if !c.stackOfOpenElements.ContainsElementInScope(t.TagName) {
				return false, inBody
			}

			c.generateImpliedEndTags()
			c.stackOfOpenElements.PopUntil(t.TagName)
		case "form":
			if len(c.containedInStackOpenElements("template")) == 0 {
				node := c.formElementPointer
				c.formElementPointer = nil
				if node == nil ||
					!c.stackOfOpenElements.ContainsElementInScope(node.NodeName) {
					return false, inBody
				}
				c.generateImpliedEndTags()
				c.stackOfOpenElements.Remove(c.stackOfOpenElements.Contains(node))
				return false, inBody
			}

			if !c.stackOfOpenElements.ContainsElementInScope("form") {
				return false, inBody
			}
			c.generateImpliedEndTags()
			c.stackOfOpenElements.PopUntil("form")
			return false, inBody
		case "p":
			if !c.stackOfOpenElements.ContainsElementInButtonScope("p") {
				c.insertHTMLElementForToken(Token{
					TagName:   "p",
					TokenType: startTagToken,
				})
			}

			c.closePElement()
		case "li":
			if !c.stackOfOpenElements.ContainsElementInListItemScope("li") {
				return false, inBody
			}
			c.generateImpliedEndTags("li")
			c.stackOfOpenElements.PopUntil("li")
		case "dd", "dt":
			if !c.stackOfOpenElements.ContainsElementInScope(t.TagName) {
				return false, inBody
			}
			c.generateImpliedEndTags(t.TagName)
			c.stackOfOpenElements.PopUntil(t.TagName)
		case "h1", "h2", "h3", "h4", "h5", "h6":
			if !c.stackOfOpenElements.ContainsElementsInScope("h1", "h2", "h3", "h4", "h5", "h6") {
				return false, inBody
			}
			c.generateImpliedEndTags()
			c.stackOfOpenElements.PopUntil("h1", "h2", "h3", "h4", "h5", "h6")
		case "sarcasm":
			//Take a deep breath, then act as described in the "any other end tag" entry below.
			c.defaultInBodyModeHandler(t)
		case "a", "b", "big", "code", "em", "font", "i", "nobr", "s", "small", "strike", "strong",
			"tt", "u":
			if c.adoptionAgencyAlgorithm(t) {
				c.defaultInBodyModeHandler(t)
			}
		case "applet", "marquee", "object":
			if !c.stackOfOpenElements.ContainsElementInScope(t.TagName) {
				return false, inBody
			}

			c.generateImpliedEndTags()
			c.stackOfOpenElements.PopUntil(t.TagName)
			c.clearListOfActiveFormattingElementsToLastMarker()
		case "br":
			t.TokenType = startTagToken
			t.Attributes = map[string]*spec.Attr{}
			return c.useRulesFor(t, inBody)
		default:
			c.defaultInBodyModeHandler(t)
		}

		return false, inBody
	case endOfFileToken:
		if len(c.stackOfTemplateInsertionModes) != 0 {
			return c.useRulesFor(t, inTemplate)
		}

		return c.stopParsing()
	}
	return false, inBody
}
func (c *HTMLTreeConstructor) textModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		c.insertCharacter(t)
		return false, text
	case endOfFileToken:
		node := c.getCurrentNode()
		if node.NodeName == "script" {
			node.AlreadyStated = true
		}
		c.stackOfOpenElements.Pop()
		return true, c.originalInsertionMode
	case endTagToken:
		switch t.TagName {
		case "script":
			c.stackOfOpenElements.Pop()
			return false, c.originalInsertionMode
		default:
			c.stackOfOpenElements.Pop()
			return false, c.originalInsertionMode
		}
	}
	return false, text
}

func (c *HTMLTreeConstructor) defaultInTableModeHandler(t Token) (bool, insertionMode) {
	c.fosterParenting = true
	repro, nextState := c.useRulesFor(t, inBody)
	c.fosterParenting = false
	return repro, nextState
}
func (c *HTMLTreeConstructor) inTableModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch c.getCurrentNode().NodeName {
		case "table", "tbody", "tfoot", "thead", "tr":
			c.pendingTableCharacterTokens = []Token{}
			c.originalInsertionMode = c.curInsertionMode
			return true, inTableText
		}
	case commentToken:
		c.insertComment(t)
		return false, inTable
	case docTypeToken:
		return false, inTable
	case startTagToken:
		switch t.TagName {
		case "caption":
			c.clearStackBackToTable()
			c.activeFormattingElements.Push(spec.ScopeMarker)
			c.insertHTMLElementForToken(t)
			return false, inCaption
		case "colgroup":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(t)
			return false, inColumnGroup
		case "col":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(Token{
				TagName:   "colgroup",
				TokenType: startTagToken,
			})
			return true, inColumnGroup
		case "tbody", "tfoot", "thead":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(t)
			return false, inTableBody
		case "td", "th", "tr":
			c.clearStackBackToTable()
			c.insertHTMLElementForToken(Token{
				TokenType: startTagToken,
				TagName:   "tbody",
			})
			return true, inTableBody
		case "table":
			repro := false
			mode := inTable
			if c.stackOfOpenElements.ContainsElementInTableScope("table") {
				c.stackOfOpenElements.PopUntil("table")
				mode = c.resetInsertionMode()
				repro = true
			}
			return repro, mode
		case "style", "script", "template":
			return c.useRulesFor(t, inHead)
		case "input":
			if attr, ok := t.Attributes["type"]; !ok || (ok && !strings.EqualFold(string(attr.Value), "hidden")) {
				return c.defaultInTableModeHandler(t)
			}

			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// ack self closing
			return false, inTable
		case "form":
			if len(c.containedInStackOpenElements("template")) != 0 ||
				c.formElementPointer != nil {
				return false, inTable
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
				return false, inTable
			}

			c.stackOfOpenElements.PopUntil("table")
			mode = c.resetInsertionMode()
			return false, mode
		case "body", "caption", "col", "colgroup", "html", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inTable
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inBody)
	}
	return c.defaultInTableModeHandler(t)
}
func (c *HTMLTreeConstructor) inTableTextModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		if t.Data == "\u0000" {
			return false, inTableText
		}
		c.pendingTableCharacterTokens = append(c.pendingTableCharacterTokens, t)
		return false, inTableText
	}

	for _, tok := range c.pendingTableCharacterTokens {
		if !isASCIIWhitespace(int(tok.Data[0])) {
			for _, t := range c.pendingTableCharacterTokens {
				c.defaultInTableModeHandler(t)
			}

			return true, c.originalInsertionMode
		}
	}

	for _, tok := range c.pendingTableCharacterTokens {
		c.insertCharacter(tok)
	}
	return true, c.originalInsertionMode
}

func (c *HTMLTreeConstructor) inCaptionHelper() {
	if !c.stackOfOpenElements.ContainsElementInTableScope("caption") {
		return
	}

	c.generateImpliedEndTags()
	c.stackOfOpenElements.PopUntil("caption")
	c.clearListOfActiveFormattingElementsToLastMarker()
}

func (c *HTMLTreeConstructor) inCaptionModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "caption", "col", "colgroup", "tbody", "td", "tfoot", "th", "thead", "tr":
			c.inCaptionHelper()
			return true, inTable
		}
	case endTagToken:
		switch t.TagName {
		case "caption":
			c.inCaptionHelper()
			return false, inTable
		case "table":
			c.inCaptionHelper()
			return true, inTable
		case "body", "col", "colgroup", "html", "tbody", "td", "tfoot", "th", "thead", "tr":
			return false, inCaption
		}
	}
	return c.useRulesFor(t, inBody)
}
func (c *HTMLTreeConstructor) inColumnGroupModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			c.insertComment(t)
			return false, inColumnGroup
		}
	case commentToken:
		c.insertComment(t)
		return false, inColumnGroup
	case docTypeToken:
		return false, inColumnGroup
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "col":
			c.insertHTMLElementForToken(t)
			c.stackOfOpenElements.Pop()
			// TODO: ack
			return false, inColumnGroup
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endTagToken:
		switch t.TagName {
		case "colgroup":
			if c.getCurrentNode().NodeName == "colgroup" {
				return false, inColumnGroup
			}
			c.stackOfOpenElements.Pop()
			return false, inTable
		case "col":
			return false, inColumnGroup
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inBody)
	}

	if c.getCurrentNode().NodeName != "colgroup" {
		return false, inColumnGroup
	}
	c.stackOfOpenElements.Pop()
	return true, inTable
}
func (c *HTMLTreeConstructor) inTableBodyModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "tr":
			c.clearStackBackToTableBody()
			c.insertHTMLElementForToken(t)
			return false, inRow
		case "th", "td":
			c.clearStackBackToTableBody()
			c.insertHTMLElementForToken(Token{
				TokenType: startTagToken,
				TagName:   "tr",
			})
			return true, inRow
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead":
			if !c.stackOfOpenElements.ContainsElementInTableScope("tbody") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("thead") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("tfoot") {
				return false, inTableBody
			}

			c.clearStackBackToTableBody()
			c.stackOfOpenElements.Pop()
			return true, inTable
		}
	case endTagToken:
		switch t.TagName {
		case "tbody", "tfoot", "thead":
			if c.stackOfOpenElements.ContainsElementInTableScope(t.TagName) {
				c.clearStackBackToTableBody()
				c.stackOfOpenElements.Pop()
				return true, inTable
			}

			return false, inTableBody
		case "table":
			if !c.stackOfOpenElements.ContainsElementInTableScope("tbody") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("thead") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("tfoot") {
				return false, inTableBody
			}
			c.clearStackBackToTableBody()
			c.stackOfOpenElements.Pop()
			return true, inTable
		case "body", "caption", "col", "colgroup", "html", "td", "th", "tr":
			return false, inTableBody
		}
	}
	return c.useRulesFor(t, inTable)
}
func (c *HTMLTreeConstructor) inRowModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "th", "td":
			c.clearStackBackToTableRow()
			c.insertHTMLElementForToken(t)
			c.activeFormattingElements.Push(spec.ScopeMarker)
			return false, inCell
		case "caption", "col", "colgroup", "tbody", "tfoot", "thead", "tr":
			if c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				c.clearStackBackToTableRow()
				c.stackOfOpenElements.Pop()
				return true, inTableBody
			}
			return false, inRow
		}
	case endTagToken:
		switch t.TagName {
		case "tr":
			if c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				c.clearStackBackToTableRow()
				c.stackOfOpenElements.Pop()
				return false, inTableBody
			}

			return false, inRow
		case "table":
			if !c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				return false, inRow
			}
			c.clearStackBackToTableRow()
			c.stackOfOpenElements.Pop()
			return true, inTableBody
		case "tbody", "tfoot", "thead":
			if !c.stackOfOpenElements.ContainsElementInTableScope(t.TagName) {
				return false, inRow
			}
			if !c.stackOfOpenElements.ContainsElementInTableScope("tr") {
				return false, inRow
			}

			c.clearStackBackToTableRow()
			c.stackOfOpenElements.Pop()
			return true, inTableBody
		case "body", "caption", "col", "colgroup", "html", "td", "th":
			return false, inRow
		}
	}
	return c.useRulesFor(t, inTable)
}

func (c *HTMLTreeConstructor) closeCell() (bool, insertionMode) {
	c.generateImpliedEndTags()
	c.stackOfOpenElements.PopUntil("td", "th")
	c.clearListOfActiveFormattingElementsToLastMarker()
	return true, inRow
}

func (c *HTMLTreeConstructor) inCellModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "caption", "col", "colgroup", "tbody", "td", "tfoot", "th", "thead", "tr":
			if !c.stackOfOpenElements.ContainsElementInTableScope("th") &&
				!c.stackOfOpenElements.ContainsElementInTableScope("td") {
				return false, inCell
			}

			return c.closeCell()
		}
	case endTagToken:
		switch t.TagName {
		case "td", "th":
			if !c.stackOfOpenElements.ContainsElementInTableScope(t.TagName) {
				return false, inCell
			}
			c.generateImpliedEndTags()
			c.stackOfOpenElements.PopUntil(t.TagName)
			c.activeFormattingElements.PopUntil(spec.ScopeMarker.NodeName)
			return false, inRow
		case "body", "caption", "col", "colgroup", "html":
			return false, inCell
		case "table", "tbody", "tfoot", "thead", "tr":
			if !c.stackOfOpenElements.ContainsElementInTableScope(t.TagName) {
				return false, inCell
			}
			return c.closeCell()
		}
	}
	return c.useRulesFor(t, inBody)
}

func (c *HTMLTreeConstructor) inSelectModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		if t.Data != "\u0000" {
			c.insertCharacter(t)
		}
	case commentToken:
		c.insertComment(t)
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
				return false, inSelect
			}

			c.stackOfOpenElements.PopUntil("select")
			return false, c.resetInsertionMode()
		case "input", "keygen", "textarea":
			if !c.stackOfOpenElements.ContainsElementInSelectScope("select") {
				return false, inSelect
			}

			c.stackOfOpenElements.PopUntil("select")
			return true, c.resetInsertionMode()
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
				return false, inSelect
			}
			c.stackOfOpenElements.Pop()
			return false, inSelect
		case "option":
			if c.getCurrentNode().NodeName == "option" {
				c.stackOfOpenElements.Pop()
				return false, inSelect
			}
			return false, inSelect
		case "select":
			if !c.stackOfOpenElements.ContainsElementInSelectScope("select") {
				return false, inSelect
			}

			c.stackOfOpenElements.PopUntil("select")
			return false, c.resetInsertionMode()
		case "template":
			return c.useRulesFor(t, inHead)
		}
	case endOfFileToken:
		return c.useRulesFor(t, inBody)
	}
	return false, inSelect
}
func (c *HTMLTreeConstructor) inSelectInTableModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case startTagToken:
		switch t.TagName {
		case "caption", "table", "tbody", "tfoot", "thead", "tr", "td", "th":
			c.stackOfOpenElements.PopUntil("select")
			return true, c.resetInsertionMode()
		}
	case endTagToken:
		switch t.TagName {
		case "caption", "table", "tbody", "tfoot", "thead", "tr", "td", "th":
			if !c.stackOfOpenElements.ContainsElementInTableScope(t.TagName) {
				return false, inSelectInTable
			}
			c.stackOfOpenElements.PopUntil("select")
			return true, c.resetInsertionMode()
		}
	}
	return c.useRulesFor(t, inSelect)
}
func (c *HTMLTreeConstructor) inTemplateModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
	case commentToken:
	case docTypeToken:
	case startTagToken:
	case endTagToken:
	case endOfFileToken:
		nodes := c.containedInStackOpenElements("template")
		if len(nodes) == 0 {
			return c.stopParsing()
		}

		// parse error
		c.stackOfOpenElements.PopUntil("template")
		c.clearListOfActiveFormattingElementsToLastMarker()
		c.stackOfTemplateInsertionModes = c.stackOfTemplateInsertionModes[:len(c.stackOfTemplateInsertionModes)-1]
		return true, c.resetInsertionMode()
	}
	return false, initial
}
func (c *HTMLTreeConstructor) afterBodyModeHandler(t Token) (bool, insertionMode) {
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
		return false, afterBody
	case docTypeToken:
		return false, afterBody
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		}
	case endTagToken:
		if t.TagName == "html" {
			if c.context != nil {
				return false, afterBody
			}
			return false, afterAfterBody
		}
	case endOfFileToken:
		c.stopParsing()
	}
	return true, inBody
}
func (c *HTMLTreeConstructor) inFramesetModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			c.insertCharacter(t)
		}
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		return false, inFrameset
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
				return false, inFrameset
			}
			c.stackOfOpenElements.Pop()
			if c.context == nil &&
				c.getCurrentNode().NodeName != "frameset" {
				return false, afterFrameset
			}
		}
	}
	return false, inFrameset
}
func (c *HTMLTreeConstructor) afterFramesetModeHandler(t Token) (bool, insertionMode) {
	switch t.TokenType {
	case characterToken:
		switch t.Data[0] {
		case '\u0009', '\u000A', '\u000C', '\u000D', '\u0020':
			c.insertCharacter(t)
		}
	case commentToken:
		c.insertComment(t)
	case docTypeToken:
		return false, afterFrameset
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
			return false, afterAfterFrameset
		}
	}
	return false, afterFrameset
}
func (c *HTMLTreeConstructor) afterAfterBodyModeHandler(t Token) (bool, insertionMode) {
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
		return false, afterAfterBody
	case docTypeToken:
		return c.useRulesFor(t, inBody)
	case endOfFileToken:
		return c.stopParsing()
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		}
	}
	return true, inBody
}
func (c *HTMLTreeConstructor) afterAfterFramesetModeHandler(t Token) (bool, insertionMode) {
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
		return c.stopParsing()
	case startTagToken:
		switch t.TagName {
		case "html":
			return c.useRulesFor(t, inBody)
		case "noframes":
			return c.useRulesFor(t, inHead)
		}
	}
	return false, afterAfterFrameset
}

func (c *HTMLTreeConstructor) switchTokenizerState(state tokenizerState) {
	c.nextTokenizerState = &state
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

type treeConstructionModeHandler func(t Token) (bool, insertionMode)

func (c *HTMLTreeConstructor) takeNextTokenizerState() *tokenizerState {
	defer func() { c.nextTokenizerState = nil }()
	return c.nextTokenizerState
}

func (c *HTMLTreeConstructor) ProcessToken(t Token) *Progress {
	reprocess := true
	for reprocess {
		reprocess, c.curInsertionMode = c.processToken(t, c.curInsertionMode)
	}
	return MakeProgress(c.getAdjustedCurrentNode(), c.takeNextTokenizerState())
}

func isMathmlIntPoint(e *spec.Node) bool {
	switch e.NodeName {
	case "mi", "mo", "mn", "ms", "mtext":
		return true
	}
	return false
}

func isHTMLIntPoint(e *spec.Node) bool {
	if e.NodeName == "annotation-xml" && e.Element.NamespaceURI == spec.Mathmlns {
		if val, ok := e.Attributes.Attrs["encoding"]; ok {
			if strings.EqualFold(string(val.Value), "text/html") || strings.EqualFold(string(val.Value), "application/xhtml+xml") {
				return true
			}
		}
	}

	if e.Element.NamespaceURI == spec.Svgns {
		if e.NodeName == "foreignObject" {
			return true
		}

		if e.NodeName == "desc" {
			return true
		}

		if e.NodeName == "title" {
			return true
		}
	}
	return false
}

var svgTable = map[string]string{
	"altglyph":            "altGlyph",
	"altglyphdef":         "altGlyphDef",
	"altglyphitem":        "altGlyphItem",
	"animatecolor":        "animateColor",
	"animatemotion":       "animateMotion",
	"animatetransform":    "animateTransform",
	"clippath":            "clipPath",
	"feblend":             "feBlend",
	"fecolormatrix":       "feColorMatrix",
	"fecomponenttransfer": "feComponentTransfer",
	"fecomposite":         "feComposite",
	"feconvolvematrix":    "feConvolveMatrix",
	"fediffuselighting":   "feDiffuseLighting",
	"fedisplacementmap":   "feDisplacementMap",
	"fedistantlight":      "feDistantLight",
	"fedropshadow":        "feDropShadow",
	"feflood":             "feFlood",
	"fefunca":             "feFuncA",
	"fefuncb":             "feFuncB",
	"fefuncg":             "feFuncG",
	"fefuncr":             "feFuncR",
	"fegaussianblur":      "feGaussianBlur",
	"feimage":             "feImage",
	"femerge":             "feMerge",
	"femergenode":         "feMergeNode",
	"femorphology":        "feMorphology",
	"feoffset":            "feOffset",
	"fepointlight":        "fePointLight",
	"fespecularlighting":  "feSpecularLighting",
	"fespotlight":         "feSpotLight",
	"fetile":              "feTile",
	"feturbulence":        "feTurbulence",
	"foreignobject":       "foreignObject",
	"glyphref":            "glyphRef",
	"lineargradient":      "linearGradient",
	"radialgradient":      "radialGradient",
	"textpath":            "textPath",
}

func (c *HTMLTreeConstructor) defaultParseTokensInForeignContentEndScriptTag(t Token, startMode insertionMode) (bool, insertionMode) {
	c.stackOfOpenElements.Pop()
	// insertion point
	// parser pause flag
	// process svg script tags
	return false, startMode
}

func (c *HTMLTreeConstructor) defaultParseTokensInForeignContentStartTag(t Token, startMode insertionMode) (bool, insertionMode) {
	acn := c.getAdjustedCurrentNode()
	switch acn.Element.NamespaceURI {
	case spec.Mathmlns:
		c.adjustMathMLAttrs(t)
	case spec.Svgns:
		if val, ok := svgTable[t.TagName]; ok {
			t.TagName = val
		}

		c.adjustSVGAttrs(t)
	}

	c.adjustForeignAttributes(t)
	c.insertForeignElementForToken(t, acn.Element.NamespaceURI)
	if t.SelfClosing {
		if t.TagName == "script" && c.getCurrentNode().Element.NamespaceURI == spec.Svgns {
			// todo: acl self closing
			return c.defaultParseTokensInForeignContentEndScriptTag(t, startMode)
		} else {
			c.stackOfOpenElements.Pop()
		}
	}
	//todo: ack self closing
	return false, startMode
}

func (c *HTMLTreeConstructor) parseTokensInForeignContent(t Token, startMode insertionMode) (bool, insertionMode) {
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

		return false, startMode
	case commentToken:
		c.insertComment(t)
		return false, startMode
	case docTypeToken:
		return false, startMode
	case startTagToken:
		switch t.TagName {
		case "b", "big", "blockquote", "body", "br", "center", "code", "dd", "div",
			"dl", "dt", "em", "embed", "h1", "h2", "h3", "h4", "h5", "h6", "head",
			"hr", "i", "img", "li", "listing", "menu", "meta", "nobr", "ol", "p",
			"pre", "ruby", "s", "small", "span", "strong", "strike", "sub", "sup",
			"table", "tt", "u", "ul", "var":
			if c.context != nil {
				return c.defaultParseTokensInForeignContentStartTag(t, startMode)
			}
			c.stackOfOpenElements.Pop()
			c.stackOfOpenElements.PopUntilConditions([]func(e *spec.Node) bool{
				isHTMLIntPoint,
				isMathmlIntPoint,
				func(e *spec.Node) bool { return e.Element != nil && e.Element.NamespaceURI == spec.Htmlns },
			})
			return true, startMode
		case "font":
			for k := range t.Attributes {
				switch k {
				case "color", "face", "size":
					if c.context != nil {
						return c.defaultParseTokensInForeignContentStartTag(t, startMode)
					}
					c.stackOfOpenElements.Pop()
					c.stackOfOpenElements.PopUntilConditions([]func(e *spec.Node) bool{
						isHTMLIntPoint,
						isMathmlIntPoint,
						func(e *spec.Node) bool { return e.Element != nil && e.Element.NamespaceURI == spec.Htmlns },
					})
					return true, startMode
				}
			}
		default:
			return c.defaultParseTokensInForeignContentStartTag(t, startMode)
		}
	case endTagToken:
		if t.TagName == "script" && c.getCurrentNode().NodeName == "svg" && c.getCurrentNode().Element.NamespaceURI == spec.Svgns {
			return c.defaultParseTokensInForeignContentEndScriptTag(t, startMode)
		}

		last := len(c.stackOfOpenElements) - 1
		for i := last; i >= 1; i-- {
			node := c.stackOfOpenElements[i]
			if i != last && node.Element.NamespaceURI == spec.Htmlns {
				return c.modeToModeHandler(startMode)(t)
			}
			if strings.EqualFold(string(node.NodeName), t.TagName) {
				for {
					popped := c.stackOfOpenElements.Pop()
					if popped == node {
						return false, startMode
					}
				}
			}
		}
	}

	return false, startMode
}

func (c *HTMLTreeConstructor) dispatch(t Token, startMode insertionMode) (bool, insertionMode) {
	acn := c.getAdjustedCurrentNode()
	if len(c.stackOfOpenElements) == 0 ||
		acn.Element.NamespaceURI == spec.Htmlns ||
		(isMathmlIntPoint(acn) && t.TokenType == startTagToken && t.TagName != "mglyph" && t.TagName != "malignmark") ||
		(isMathmlIntPoint(acn) && t.TokenType == characterToken) ||
		(acn.NodeName == "annotation-xml" && t.TokenType == startTagToken && t.TagName == "svg") ||
		(isHTMLIntPoint(acn) && (t.TokenType == startTagToken || t.TokenType == characterToken)) {
		return c.modeToModeHandler(startMode)(t)
	}

	return c.parseTokensInForeignContent(t, startMode)
}

func (c *HTMLTreeConstructor) processToken(t Token, startMode insertionMode) (bool, insertionMode) {
	fmt.Printf("[TREE]token: %+vmode: %s\n", t, startMode)
	reprocess, nextMode := c.dispatch(t, startMode)
	fmt.Printf("[TREE]tree after: \n%s\n\n", c.HTMLDocument.Node)
	return reprocess, nextMode
}
