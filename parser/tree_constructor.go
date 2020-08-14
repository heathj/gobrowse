package parser

import "strings"

// Element is an individual HTML element that gets added to the DOM.
type Element struct {
	elemType elementType
	children []*Element
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
	documentElement
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

type documentType uint

const (
	iframeSrcDoc documentType = iota
)

// HTMLTreeConstructor holds the state for various state of the tree construction phase.
type HTMLTreeConstructor struct {
	tokenChannel             chan *Token
	config                   htmlParserConfig
	originalInsertionMode    insertionMode
	scriptingEnabled         bool
	document                 *Document
	framesetOK               bool
	openElements             []*Element
	headPointer              *Element
	formPointer              *Element
	activeFormattingElements []formattingElement
	mappings                 map[insertionMode]treeConstructionModeHandler
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

type location uint

const (
	lastChildOfDocument location = iota
	adjustedInsertionLocation
)

// NewHTMLTreeConstructor creates an HTMLTreeConstructor.
func NewHTMLTreeConstructor() *HTMLTreeConstructor {
	c := make(chan *Token, 10)
	tr := &HTMLTreeConstructor{
		tokenChannel: c,
		document:     &Document{},
	}

	tr.createMappings()
	return tr
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

func (c *HTMLTreeConstructor) SearchOpenElements(e elementType) bool {
	for _, oe := range c.openElements {
		if oe.elemType == e {
			return true
		}
	}
	return false
}

// Inserts a comment at a specific location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertCommentAt(t *Token, l location) {
	//TODO
}

// Inserts a comment at the adjusted insertion location.
// https://html.spec.whatwg.org/multipage/parsing.html#insert-a-comment
func (c *HTMLTreeConstructor) insertComment(t *Token) {
	c.insertCommentAt(t, adjustedInsertionLocation)
}

// https://html.spec.whatwg.org/multipage/parsing.html#appropriate-place-for-inserting-a-node
func (c *HTMLTreeConstructor) getAppropriatePlaceForInsertion() int {
	//TODO
	return 0
}

func (c *HTMLTreeConstructor) WriteMarker() {
	c.activeFormattingElements = append(c.activeFormattingElements, markerFElement)
}

// createElementForToken creates an element from a token with the provided
// namespace and parent element.
// https://html.spec.whatwg.org/multipage/parsing.html#create-an-element-for-the-token
func (c *HTMLTreeConstructor) createElementForToken(t *Token, ns namespace, ip *Element) *Element {
	e := &Element{}
	c.document.children = append(c.document.children, e)
	return e
}

//WriteHTMLElement inserts a foreign element for the token, in the HTML namespace.
func (c *HTMLTreeConstructor) WriteHTMLElement(t *Token) {

}

//WriteCharacter adds the given characer to the appropriate place for inserting a node.
func (c *HTMLTreeConstructor) WriteCharacter(t *Token) {

}

//WriteLatestHeadElement sets the head element pointer to the newly created head element.
func (c *HTMLTreeConstructor) WriteLatestHeadElement() {

}

// PushOpenElements pushes an element to the list of currently open elements being parsed.
func (c *HTMLTreeConstructor) PushOpenElements(e *Element) {
	c.openElements = append(c.openElements, e)
}

// PopOpenElements pops an element off the list of open elements
func (c *HTMLTreeConstructor) PopOpenElements() {

}

// WriteDocumentType sets the current document type given the curren token.
func (c *HTMLTreeConstructor) WriteDocumentType(t *Token) {}

// Quirks sets the quirks value to "force".
func (c *HTMLTreeConstructor) Quirks() {
	c.quirks = force
}

// LimitedQuirks sets the quirks value to "limited"
func (c *HTMLTreeConstructor) LimitedQuirks() {
	c.quirks = limited
}

// FramesetNotOK sets framesetOK state to false.
func (c *HTMLTreeConstructor) FramesetNotOK() {
	c.framesetOK = false
}

// CurrentNode returns the bottommost node from the stack of open elements.
func (c *HTMLTreeConstructor) CurrentNode() *Element {
	return c.openElements[len(c.openElements)-1]
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

func (c *HTMLTreeConstructor) reconstructActiveFormattingElements() {

}

func (c *HTMLTreeConstructor) clearStackBackToTable() {
	for {
		switch c.CurrentNode().elemType {
		case tableElement, templateElement, htmlElement:
			return
		default:
			c.PopOpenElements()
		}
	}
}

func (c *HTMLTreeConstructor) openElementsInTableScope(elem elementType) bool {
	return false
}

func (c *HTMLTreeConstructor) resetInsertionMode() insertionMode {

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

func (c *HTMLTreeConstructor) isForceQuirks(t *Token) bool {
	if c.document.docType != iframeSrcDoc {
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
			t.SystemIdentifier != missing ||
			t.SystemIdentifier != "about:legacy-compat" {
			//TODO: just says this was a parse error?
			err = generalParseError
		}
		c.WriteDocumentType(t)

		if c.isForceQuirks(t) {
			c.Quirks()
		} else if c.isLimitedQuirks(t) {
			c.LimitedQuirks()
		}

		return false, beforeHTML, err
	default:
		if c.document.docType != iframeSrcDoc {
			//TODO: this is a parse error?
			err = generalParseError
			c.Quirks()
		}
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
		c.insertCommentAt(t, lastChildOfDocument)
		return false, beforeHTML, noError
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			return false, beforeHTML, noError
		}
	case startTagToken:
		if t.TagName == "html" {
			elem := c.createElementForToken(t, htmlNamepsace, c.document.elem)
			c.document.elem.children = append(c.document.elem.children, elem)
			c.openElements = append(c.openElements, elem)
		}
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
	c.WriteHTMLElement(t)
	c.WriteLatestHeadElement()
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
			c.WriteHTMLElement(t)
			// TODO: set head element pointer
			c.WriteLatestHeadElement()
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
			c.WriteHTMLElement(t)
			c.PopOpenElements()
			//TODO: acknowledge the self closing flag?
		case "meta":
			c.WriteHTMLElement(t)
			c.PopOpenElements()
			//TODO: acknowledge the self closing flag?
		case "title":
		case "noscript":
			if c.scriptingEnabled {

			} else {
				c.WriteHTMLElement(t)
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
			c.PopOpenElements()
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
	c.PopOpenElements()
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
			c.PopOpenElements()
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
	c.WriteHTMLElement(&Token{
		TokenType: startTagToken,
		TagName:   "body",
	})
	return true, inBody, noError
}
func (c *HTMLTreeConstructor) afterHeadModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		switch t.Data {
		case "\u0009", "\u000A", "\u000C", "\u000D", "\u0020":
			c.WriteCharacter(t)
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
			c.WriteHTMLElement(t)
			c.FramesetNotOK()
			return false, inBody, noError
		case "frameset":
			c.WriteHTMLElement(t)
			return false, inFrameset, noError
		case "base", "basefont", "bgsound", "link", "meta", "noframes", "script", "style", "template", "title":
			c.PushOpenElements(c.headPointer)
			reprocess, nextmode, err := c.inHeadModeHandler(t)

			c.PopOpenElements()
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
			c.WriteCharacter(t)
			return false, inBody, noError
		default:
			c.reconstructActiveFormattingElements()
			c.WriteCharacter(t)
			c.FramesetNotOK()
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
		case "applet", "marquee", "object":
		case "br":
		default:

		}

		return false, inBody, err
	case endOfFileToken:
	}
	return false, initial, noError
}
func (c *HTMLTreeConstructor) textModeHandler(t *Token) (bool, insertionMode, parseError) {
	switch t.TokenType {
	case characterToken:
		c.WriteCharacter(t)
		return false, text, noError
	case endOfFileToken:
		return true, c.originalInsertionMode, generalParseError
	case endTagToken:
		switch t.TagName {
		case "script":
			return false, c.originalInsertionMode, noError
		default:
			c.PopOpenElements()
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
		switch c.CurrentNode().elemType {
		case tableElement, tbodyElement, tfootElement, theadElement, trElement:
			c.originalInsertionMode = inTable
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
			c.WriteMarker()
			c.WriteHTMLElement(t)
			return false, inColumnGroup, noError
		case "colgroup":
			c.clearStackBackToTable()
			c.WriteHTMLElement(t)
			return false, inColumnGroup, noError
		case "col":
			c.clearStackBackToTable()
			c.WriteHTMLElement(&Token{
				TagName:   "colgroup",
				TokenType: startTagToken,
			})
			return true, inColumnGroup, noError
		case "tbody", "tfoot", "thead":
			c.clearStackBackToTable()
			c.WriteHTMLElement(t)
			return false, inTableBody, noError
		case "td", "th", "tr":
			c.clearStackBackToTable()
			c.WriteHTMLElement(&Token{
				TokenType: startTagToken,
				TagName:   "tbody",
			})
			return true, inTableBody, noError
		case "table":
			if c.openElementsInTableScope(tableElement) {
				return false, inTable, noError
			}

			c.clearStackBackToTable()
			c.PopOpenElements()
			nextMode := c.resetInsertionMode()
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

			c.WriteHTMLElement(t)
			c.PopOpenElements()
			//TODO self closing
			return false, inTable, generalParseError
		case "form":
			if c.SearchOpenElements(templateElement) || c.formPointer != nil {
				return false, inTable, generalParseError
			}

			c.WriteHTMLElement(t)
			//			c.formPointer =
			c.PopOpenElements()
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
