package parser

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"sync"
)

type htmlParserConfigKey uint
type htmlParserConfig map[htmlParserConfigKey]int

const (
	debug = iota
)

// HTMLTokenizer holds state for the various state of the tokenizer.
type HTMLTokenizer struct {
	eof                     bool
	config                  htmlParserConfig
	returnState             tokenizerState
	htmlReader              *bufio.Reader
	tokenChannel            chan *Token
	wg                      sync.WaitGroup
	mappings                map[tokenizerState]parserStateHandler
	tokenBuilder            *TokenBuilder
	lastEmittedStartTagName string
}

// NewHTMLTokenizer creates an HTML parser that can be used to process
// an HTML string.
func NewHTMLTokenizer(htmlStr string, config htmlParserConfig) (*HTMLTokenizer, chan *Token, *sync.WaitGroup) {
	tokenChannel := make(chan *Token, 10)

	p := &HTMLTokenizer{
		returnState:  dataState,
		config:       config,
		htmlReader:   bufio.NewReader(strings.NewReader(htmlStr)),
		tokenChannel: tokenChannel,
		tokenBuilder: newTokenBuilder(),
	}

	createMappings(p)

	return p, tokenChannel, &p.wg
}

func createMappings(p *HTMLTokenizer) {

	// doing the mappings like this allows us to not have to use nested recursion when following the
	// state machine and also makes testing individual states simpler.
	p.mappings = map[tokenizerState]parserStateHandler{
		dataState:                                     p.dataStateParser,
		rcDataState:                                   p.rcDataStateParser,
		rawTextState:                                  p.rawTextStateParser,
		scriptDataState:                               p.scriptDataStateParser,
		plaintextState:                                p.plaintextStateParser,
		tagOpenState:                                  p.tagOpenStateParser,
		endTagOpenState:                               p.endTagOpenStateParser,
		tagNameState:                                  p.tagNameStateParser,
		rcDataLessThanSignState:                       p.rcDataLessThanSignStateParser,
		rcDataEndTagOpenState:                         p.rcDataEndTagOpenStateParser,
		rcDataEndTagNameState:                         p.rcDataEndTagNameStateParser,
		rawTextLessThanSignState:                      p.rawTextLessThanSignStateParser,
		rawTextEndTagOpenState:                        p.rawTextEndTagOpenStateParser,
		rawTextEndTagNameState:                        p.rawTextEndTagNameStateParser,
		scriptDataLessThanSignState:                   p.scriptDataLessThanSignStateParser,
		scriptDataEndTagOpenState:                     p.scriptDataEndTagOpenStateParser,
		scriptDataEndTagNameState:                     p.scriptDataEndTagNameStateParser,
		scriptDataEscapeStartState:                    p.scriptDataEscapeStartStateParser,
		scriptDataEscapeStartDashState:                p.scriptDataEscapeStartDashStateParser,
		scriptDataEscapedState:                        p.scriptDataEscapedStateParser,
		scriptDataEscapedDashState:                    p.scriptDataEscapedDashStateParser,
		scriptDataEscapedDashDashState:                p.scriptDataEscapedDashDashStateParser,
		scriptDataEscapedLessThanSignState:            p.scriptDataEscapedLessThanSignStateParser,
		scriptDataEscapedEndTagOpenState:              p.scriptDataEscapedEndTagOpenStateParser,
		scriptDataEscapedEndTagNameState:              p.scriptDataEscapedEndTagNameStateParser,
		scriptDataDoubleEscapeStartState:              p.scriptDataDoubleEscapeStartStateParser,
		scriptDataDoubleEscapedState:                  p.scriptDataDoubleEscapedStateParser,
		scriptDataDoubleEscapedDashState:              p.scriptDataDoubleEscapedDashStateParser,
		scriptDataDoubleEscapedDashDashState:          p.scriptDataDoubleEscapedDashDashStateParser,
		scriptDataDoubleEscapedLessThanSignState:      p.scriptDataDoubleEscapedLessThanSignStateParser,
		scriptDataDoubleEscapeEndState:                p.scriptDataDoubleEscapeEndStateParser,
		beforeAttributeNameState:                      p.beforeAttributeNameStateParser,
		attributeNameState:                            p.attributeNameStateParser,
		afterAttributeNameState:                       p.afterAttributeNameStateParser,
		beforeAttributeValueState:                     p.beforeAttributeValueStateParser,
		attributeValueDoubleQuotedState:               p.attributeValueDoubleQuotedStateParser,
		attributeValueSingleQuotedState:               p.attributeValueSingleQuotedStateParser,
		attributeValueUnquotedState:                   p.attributeValueUnquotedStateParser,
		afterAttributeValueQuotedState:                p.afterAttributeValueQuotedStateParser,
		selfClosingStartTagState:                      p.selfClosingStartTagStateParser,
		bogusCommentState:                             p.bogusCommentStateParser,
		markupDeclarationOpenState:                    p.markupDeclarationOpenStateParser,
		commentStartState:                             p.commentStartStateParser,
		commentStartDashState:                         p.commentStartDashStateParser,
		commentState:                                  p.commentStateParser,
		commentLessThanSignState:                      p.commentLessThanSignStateParser,
		commentLessThanSignBangState:                  p.commentLessThanSignBangStateParser,
		commentLessThanSignBangDashState:              p.commentLessThanSignBangDashStateParser,
		commentLessThanSignBangDashDashState:          p.commentLessThanSignBangDashDashStateParser,
		commentEndDashState:                           p.commentEndDashStateParser,
		commentEndState:                               p.commentEndStateParser,
		commentEndBangState:                           p.commentEndBangStateParser,
		doctypeState:                                  p.doctypeStateParser,
		beforeDoctypeNameState:                        p.beforeDoctypeNameStateParser,
		doctypeNameState:                              p.doctypeNameStateParser,
		afterDoctypeNameState:                         p.afterDoctypeNameStateParser,
		afterDoctypePublicKeywordState:                p.afterDoctypePublicKeywordStateParser,
		beforeDoctypePublicIdentifierState:            p.beforeDoctypePublicIdentifierStateParser,
		doctypePublicIdentifierDoubleQuotedState:      p.doctypePublicIdentifierDoubleQuotedStateParser,
		doctypePublicIdentifierSingleQuotedState:      p.doctypePublicIdentifierSingleQuotedStateParser,
		afterDoctypePublicIdentifierState:             p.afterDoctypePublicIdentifierStateParser,
		betweenDoctypePublicAndSystemIdentifiersState: p.betweenDoctypePublicAndSystemIdentifiersStateParser,
		afterDoctypeSystemKeywordState:                p.afterDoctypeSystemKeywordStateParser,
		beforeDoctypeSystemIdentifierState:            p.beforeDoctypeSystemIdentifierStateParser,
		doctypeSystemIdentifierDoubleQuotedState:      p.doctypeSystemIdentifierDoubleQuotedStateParser,
		doctypeSystemIdentifierSingleQuotedState:      p.doctypeSystemIdentifierSingleQuotedStateParser,
		afterDoctypeSystemIdentifierState:             p.afterDoctypeSystemIdentifierStateParser,
		bogusDoctypeState:                             p.bogusDoctypeStateParser,
		cdataSectionState:                             p.cdataSectionStateParser,
		cdataSectionBracketState:                      p.cdataSectionBracketStateParser,
		cdataSectionEndState:                          p.cdataSectionEndStateParser,
		characterReferenceState:                       p.characterReferenceStateParser,
		namedCharacterReferenceState:                  p.namedCharacterReferenceStateParser,
		ambiguousAmpersandState:                       p.ambiguousAmpersandStateParser,
		numericCharacterReferenceState:                p.numericCharacterReferenceStateParser,
		hexadecimalCharacterReferenceStartState:       p.hexadecimalCharacterReferenceStartStateParser,
		decimalCharacterReferenceStartState:           p.decimalCharacterReferenceStartStateParser,
		hexadecimalCharacterReferenceState:            p.hexadecimalCharacterReferenceStateParser,
		decimalCharacterReferenceState:                p.decimalCharacterReferenceStateParser,
		numericCharacterReferenceEndState:             p.numericCharacterReferenceEndStateParser,
	}
}

// TODO: implement
func (p *HTMLTokenizer) adjustedCurrentNode() bool {
	return false
}

// TODO: implement
func (p *HTMLTokenizer) inHTMLNamepsace() bool {
	return false
}

func isNonCharacter(code int) bool {
	if code >= 0xFDD0 && code <= 0xFDEF {
		return true
	}

	switch code {
	case 0xFFFE, 0xFFFF, 0x1FFFE, 0x1FFFF, 0x2FFFE, 0x2FFFF, 0x3FFFE, 0x3FFFF, 0x4FFFE, 0x4FFFF, 0x5FFFE, 0x5FFFF, 0x6FFFE, 0x6FFFF, 0x7FFFE, 0x7FFFF, 0x8FFFE, 0x8FFFF, 0x9FFFE, 0x9FFFF, 0xAFFFE, 0xAFFFF, 0xBFFFE, 0xBFFFF, 0xCFFFE, 0xCFFFF, 0xDFFFE, 0xDFFFF, 0xEFFFE, 0xEFFFF, 0xFFFFE, 0xFFFFF, 0x10FFFE, 0x10FFFF:
		return true
	default:
		return false
	}
}

func isC0Control(code int) bool {
	if code >= 0x00 && code <= 0x1F {
		return true
	}

	return false
}

func isControl(code int) bool {
	if isC0Control(code) || (code >= 0x7F && code <= 0x9F) {
		return true
	}

	return false
}

func isASCIIWhitespace(code int) bool {
	switch code {
	case 0x09, 0x0A, 0x0C, 0x0D, 0x20:
		return true
	default:
		return false
	}
}

func isSurrogate(code int) bool {
	if code >= 0xD800 && code <= 0xDFFF {
		return true
	}
	return false
}

func (p *HTMLTokenizer) wasConsumedByAttribute() bool {
	if p.returnState == attributeValueDoubleQuotedState ||
		p.returnState == attributeValueSingleQuotedState ||
		p.returnState == attributeValueUnquotedState {
		return true
	}
	return false
}

func (p *HTMLTokenizer) flushCodePointsAsCharacterReference() {
	if p.wasConsumedByAttribute() {
		for _, v := range p.tokenBuilder.TempBuffer() {
			p.tokenBuilder.WriteAttributeValue(v)
		}
	} else {
		for _, v := range p.tokenBuilder.TempBuffer() {
			p.emit(p.tokenBuilder.CharacterToken(v))
		}
	}
}

func (p *HTMLTokenizer) isApprEndTagToken() bool {
	return p.lastEmittedStartTagName == p.tokenBuilder.name.String()
}

func (p *HTMLTokenizer) emit(tok Token) {
	if tok.TokenType == endTagToken {
		// When an end tag token is emitted with attributes, that is an end-tag-with-attributes
		// parse error.
		if len(tok.Attributes) > 0 {
			logError(endTagWithAttributes)
			tok.Attributes = make(map[string]string)
		}

		// When an end tag token is emitted with its self-closing flag set, that is an
		// end-tag-with-trailing-solidus parse error.
		if tok.SelfClosing {
			logError(endTagWithTrailingSolidus)
			tok.SelfClosing = false
		}
	} else if tok.TokenType == startTagToken {
		p.lastEmittedStartTagName = tok.TagName
	}
	p.tokenChannel <- &tok
}

func (p *HTMLTokenizer) isEndOfFile() bool {
	return p.eof
}

func (p *HTMLTokenizer) dataStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '&':
		p.returnState = dataState
		return false, characterReferenceState, noError
	case '<':
		return false, tagOpenState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, dataState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, dataState, noError
	}
}

func (p *HTMLTokenizer) rcDataStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '&':
		p.returnState = rcDataState
		return false, characterReferenceState, noError
	case '<':
		return false, rcDataLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, rcDataState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, rcDataState, noError
	}
}
func (p *HTMLTokenizer) rawTextStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '<':
		return false, rawTextLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, rawTextState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, rawTextState, noError
	}
}
func (p *HTMLTokenizer) scriptDataStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '<':
		return false, scriptDataLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataState, noError
	}
}

func (p *HTMLTokenizer) plaintextStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, plaintextState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, plaintextState, noError
	}
}

func (p *HTMLTokenizer) tagOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofBeforeTagName
	}
	switch r {
	case '!':
		return false, markupDeclarationOpenState, noError
	case '/':
		return false, endTagOpenState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.curTagType = startTag
		return true, tagNameState, noError
	case '?':
		p.tokenBuilder.NewToken()
		return true, bogusCommentState, unexpectedQuestionMakrInsteadofTagName
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, dataState, invalidFirstCharacterOfTagName
	}
}

func (p *HTMLTokenizer) endTagOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.CharacterToken('/'))
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofBeforeTagName
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.curTagType = endTag
		return true, tagNameState, noError
	case '>':

		return false, dataState, missingEndTagName
	default:
		p.tokenBuilder.NewToken()
		return true, bogusCommentState, invalidFirstCharacterOfTagName
	}
}

func (p *HTMLTokenizer) tagNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020': // tab, line feed, form feed, space
		return false, beforeAttributeNameState, noError
	case '/':
		return false, selfClosingStartTagState, noError
	case '>':
		p.emitCurrentTag()
		return false, dataState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, tagNameState, noError
	case '\u0000': // null
		p.tokenBuilder.WriteName('\uFFFD')
		return false, tagNameState, unexpectedNullCharacter
	default:
		p.tokenBuilder.WriteName(r)
		return false, tagNameState, noError
	}
}

func (p *HTMLTokenizer) rcDataLessThanSignStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rcDataState, noError
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, rcDataEndTagOpenState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rcDataState, noError
	}
}

func (p *HTMLTokenizer) defaultRcDataEndTagOpenStateParser() (bool, tokenizerState, parseError) {
	p.emit(p.tokenBuilder.CharacterToken('<'))
	p.emit(p.tokenBuilder.CharacterToken('/'))
	return true, rcDataState, noError
}
func (p *HTMLTokenizer) rcDataEndTagOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return p.defaultRcDataEndTagOpenStateParser()
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.curTagType = endTag
		return true, rcDataEndTagNameState, noError
	default:
		return p.defaultRcDataEndTagOpenStateParser()
	}
}

func (p *HTMLTokenizer) defaultRcDataEndTagNameStateCase() (bool, tokenizerState, parseError) {
	p.emit(p.tokenBuilder.CharacterToken('<'))
	p.emit(p.tokenBuilder.CharacterToken('/'))
	for _, r := range p.tokenBuilder.TempBuffer() {
		p.emit(p.tokenBuilder.CharacterToken(r))
	}
	return true, rcDataState, noError
}
func (p *HTMLTokenizer) rcDataEndTagNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return p.defaultRcDataEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState, noError
		}
		return p.defaultRcDataEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState, noError
		}
		return p.defaultRcDataEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			p.emitCurrentTag()
			return false, dataState, noError
		}
		return p.defaultRcDataEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, rcDataEndTagNameState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, rcDataEndTagNameState, noError
	default:
		return p.defaultRcDataEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) rawTextLessThanSignStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rawTextState, noError
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, rawTextEndTagOpenState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rawTextState, noError
	}
}

func (p *HTMLTokenizer) defaultRawTextEndTagOpenStateParser() (bool, tokenizerState, parseError) {
	p.emit(p.tokenBuilder.CharacterToken('<'))
	p.emit(p.tokenBuilder.CharacterToken('/'))
	return true, rawTextState, noError
}

func (p *HTMLTokenizer) rawTextEndTagOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return p.defaultRawTextEndTagOpenStateParser()
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.curTagType = endTag
		return true, rawTextEndTagNameState, noError
	default:
		return p.defaultRawTextEndTagOpenStateParser()
	}
}

func (p *HTMLTokenizer) defaultRawTextEndTagNameStateCase() (bool, tokenizerState, parseError) {
	p.emit(p.tokenBuilder.CharacterToken('<'))
	p.emit(p.tokenBuilder.CharacterToken('/'))
	for _, r := range p.tokenBuilder.TempBuffer() {
		p.emit(p.tokenBuilder.CharacterToken(r))
	}
	return true, rawTextState, noError
}
func (p *HTMLTokenizer) rawTextEndTagNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return p.defaultRawTextEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState, noError
		}
		return p.defaultRawTextEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState, noError
		}
		return p.defaultRawTextEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			p.emitCurrentTag()
			return false, dataState, noError
		}
		return p.defaultRawTextEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, rawTextEndTagNameState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, rawTextEndTagNameState, noError
	default:
		return p.defaultRawTextEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) scriptDataLessThanSignStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataState, noError
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, scriptDataEndTagOpenState, noError
	case '!':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.CharacterToken('!'))
		return false, scriptDataEscapeStartState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataState, noError
	}
}

func (p *HTMLTokenizer) scriptDataEndTagOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataState, noError
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.curTagType = endTag
		return true, scriptDataEndTagNameState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataState, noError
	}
}

func (p *HTMLTokenizer) defaultScriptDataEndTagNameStateCase() (bool, tokenizerState, parseError) {
	p.emit(p.tokenBuilder.CharacterToken('<'))
	p.emit(p.tokenBuilder.CharacterToken('/'))
	for _, r := range p.tokenBuilder.TempBuffer() {
		p.emit(p.tokenBuilder.CharacterToken(r))
	}
	return true, scriptDataState, noError
}
func (p *HTMLTokenizer) scriptDataEndTagNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return p.defaultScriptDataEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState, noError
		}
		return p.defaultScriptDataEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState, noError
		}
		return p.defaultScriptDataEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			p.emitCurrentTag()
			return false, dataState, noError
		}
		return p.defaultScriptDataEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEndTagNameState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEndTagNameState, noError
	default:
		return p.defaultScriptDataEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) scriptDataEscapeStartStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, scriptDataState, noError
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapeStartDashState, noError
	default:
		return true, scriptDataState, noError
	}
}
func (p *HTMLTokenizer) scriptDataEscapeStartDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, scriptDataState, noError
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashDashState, noError
	default:
		return true, scriptDataState, noError
	}
}
func (p *HTMLTokenizer) scriptDataEscapedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInScriptHTMLCommentLikeText
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashState, noError
	case '<':
		return false, scriptDataEscapedLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataEscapedState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataEscapedDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInScriptHTMLCommentLikeText
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashDashState, noError
	case '<':
		return false, scriptDataEscapedLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataEscapedState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataEscapedDashDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInScriptHTMLCommentLikeText
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashDashState, noError
	case '<':
		return false, scriptDataEscapedLessThanSignState, noError
	case '>':
		p.emit(p.tokenBuilder.CharacterToken('>'))
		return false, scriptDataState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataEscapedState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataEscapedLessThanSignStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataEscapedState, noError
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, scriptDataEscapedEndTagOpenState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.ResetTempBuffer()
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataDoubleEscapeStartState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataEscapedEndTagOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataEscapedState, noError
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.curTagType = endTag
		return true, scriptDataEscapedEndTagNameState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		p.emit(p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataEscapedState, noError
	}
}

func (p *HTMLTokenizer) defaultScriptDataEscapedEndTagNameStateCase() (bool, tokenizerState, parseError) {
	p.emit(p.tokenBuilder.CharacterToken('<'))
	p.emit(p.tokenBuilder.CharacterToken('/'))
	for _, r := range p.tokenBuilder.TempBuffer() {
		p.emit(p.tokenBuilder.CharacterToken(r))
	}
	return true, scriptDataEscapedState, noError
}
func (p *HTMLTokenizer) scriptDataEscapedEndTagNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState, noError
		}
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState, noError
		}
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			p.emitCurrentTag()
			return false, dataState, noError
		}
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEscapedEndTagNameState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEscapedEndTagNameState, noError
	default:
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapeStartStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, scriptDataEscapedState, noError
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020', '/', '>':
		p.emit(p.tokenBuilder.CharacterToken(r))
		if p.tokenBuilder.TempBuffer() == "script" {
			return false, scriptDataDoubleEscapedState, noError
		}
		return false, scriptDataEscapedState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		r += 0x20
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeStartState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeStartState, noError
	default:
		return true, scriptDataEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInScriptHTMLCommentLikeText
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataDoubleEscapedDashState, noError
	case '<':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return false, scriptDataDoubleEscapedLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataDoubleEscapedState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataDoubleEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInScriptHTMLCommentLikeText
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataDoubleEscapedDashDashState, noError
	case '<':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return false, scriptDataDoubleEscapedLessThanSignState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataDoubleEscapedState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataDoubleEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedDashDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInScriptHTMLCommentLikeText
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataDoubleEscapedDashDashState, noError
	case '<':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return false, scriptDataDoubleEscapedLessThanSignState, noError
	case '>':
		p.emit(p.tokenBuilder.CharacterToken('>'))
		return false, scriptDataState, noError
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataDoubleEscapedState, unexpectedNullCharacter
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataDoubleEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedLessThanSignStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, scriptDataDoubleEscapedState, noError
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		p.emit(p.tokenBuilder.CharacterToken('/'))
		return false, scriptDataDoubleEscapeEndState, noError
	default:
		return true, scriptDataDoubleEscapedState, noError
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapeEndStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, scriptDataDoubleEscapedState, noError
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020', '/', '>':
		p.emit(p.tokenBuilder.CharacterToken(r))
		if p.tokenBuilder.TempBuffer() == "script" {
			return false, scriptDataEscapedState, noError
		}
		return false, scriptDataDoubleEscapedState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		r += 0x20
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeEndState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeEndState, noError
	default:
		return true, scriptDataDoubleEscapedState, noError
	}
}

func (p *HTMLTokenizer) beforeAttributeNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, afterAttributeNameState, noError
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeAttributeNameState, noError
	case '/', '>':
		return true, afterAttributeNameState, noError
	case '=':
		// set that attribute's name to the current input character, and its value to the empty string.
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState, unexpectedEqualsSignBeforeAttributeName
	default:
		return true, attributeNameState, noError
	}
}

func (p *HTMLTokenizer) attributeNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.CommitAttribute()
		return true, afterAttributeNameState, noError
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020', '/', '>':
		p.tokenBuilder.CommitAttribute()
		return true, afterAttributeNameState, noError
	case '=':
		if p.tokenBuilder.RemoveDuplicateAttributeName() {
			return false, beforeAttributeValueState, duplicateAttribute
		}
		return false, beforeAttributeValueState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		r += 0x20
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState, noError
	case '\u0000':
		p.tokenBuilder.WriteAttributeName('\uFFFD')
		return false, attributeNameState, unexpectedNullCharacter
	case '"', '\'', '<':
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState, unexpectedCharacterInAttributeName
	default:
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState, noError
	}
}

func (p *HTMLTokenizer) afterAttributeNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterAttributeNameState, noError
	case '/':
		return false, selfClosingStartTagState, noError
	case '=':
		return false, beforeAttributeValueState, noError
	case '>':
		p.emitCurrentTag()
		return false, dataState, noError
	default:
		return true, attributeNameState, noError
	}
}

func (p *HTMLTokenizer) beforeAttributeValueStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, attributeValueUnquotedState, noError
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeAttributeValueState, noError
	case '"':
		return false, attributeValueDoubleQuotedState, noError
	case '\'':
		return false, attributeValueSingleQuotedState, noError
	case '>':
		p.tokenBuilder.CommitAttribute()
		p.emitCurrentTag()
		return false, dataState, missingAttributeValue
	default:
		return true, attributeValueUnquotedState, noError
	}
}

func (p *HTMLTokenizer) attributeValueDoubleQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '"':
		p.tokenBuilder.CommitAttribute()
		return false, afterAttributeValueQuotedState, noError
	case '&':
		p.returnState = attributeValueDoubleQuotedState
		return false, characterReferenceState, noError
	case '\u0000':
		p.tokenBuilder.WriteAttributeValue('\uFFFD')
		return false, attributeValueDoubleQuotedState, unexpectedNullCharacter
	default:
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueDoubleQuotedState, noError
	}
}

func (p *HTMLTokenizer) attributeValueSingleQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '\'':

		p.tokenBuilder.CommitAttribute()
		return false, afterAttributeValueQuotedState, noError
	case '&':
		p.returnState = attributeValueSingleQuotedState
		return false, characterReferenceState, noError
	case '\u0000':
		p.tokenBuilder.WriteAttributeValue('\uFFFD')
		return false, attributeValueSingleQuotedState, unexpectedNullCharacter
	default:
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueSingleQuotedState, noError
	}
}

func (p *HTMLTokenizer) attributeValueUnquotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		p.tokenBuilder.CommitAttribute()
		return false, beforeAttributeNameState, noError
	case '&':
		p.returnState = attributeValueUnquotedState
		return false, characterReferenceState, noError
	case '>':
		p.tokenBuilder.CommitAttribute()
		p.emitCurrentTag()
		return false, dataState, noError
	case '\u0000':
		p.tokenBuilder.WriteAttributeValue('\uFFFD')
		return false, attributeValueUnquotedState, unexpectedNullCharacter
	case '"', '\'', '<', '=', '`':
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueUnquotedState, unexpectedCharacterInUnquotedAttributeValue
	default:
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueUnquotedState, noError
	}
}

func (p *HTMLTokenizer) afterAttributeValueQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeAttributeNameState, noError
	case '/':
		return false, selfClosingStartTagState, noError
	case '>':
		p.emitCurrentTag()
		return false, dataState, noError
	default:
		return true, beforeAttributeNameState, missingWhitespaceBetweenAttributes
	}
}

func (p *HTMLTokenizer) selfClosingStartTagStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInTag
	}
	switch r {
	case '>':
		p.tokenBuilder.EnableSelfClosing()
		p.emitCurrentTag()
		return false, dataState, noError
	default:
		return true, beforeAttributeNameState, unexpectedSolidusInTag
	}
}

func (p *HTMLTokenizer) bogusCommentStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CommentToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState, noError
	case '\u0000':
		p.tokenBuilder.WriteData('\uFFFD')
		return false, bogusCommentState, unexpectedNullCharacter
	default:
		p.tokenBuilder.WriteData(r)
		return false, bogusCommentState, noError
	}
}

// used below to look for peeking at what state to jump to next
var doctype = []byte("octype")
var cdata = []byte("CDATA[")
var peekDist = 6

func (p *HTMLTokenizer) defaultMarkupDeclarationOpenStateParser() (bool, tokenizerState, parseError) {
	p.tokenBuilder.NewToken()
	return true, bogusCommentState, incorrectlyOpenedComment
}

func (p *HTMLTokenizer) markupDeclarationOpenStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.NewToken()
		return true, bogusCommentState, incorrectlyOpenedComment
	}
	var (
		peeked []byte
		err    error
	)

	switch r {
	case '-':
		peeked, err = p.htmlReader.Peek(1)
		if err != nil {
			if len(peeked) < 1 {
				return p.defaultMarkupDeclarationOpenStateParser()
			}
		}
		if len(peeked) == 1 && peeked[0] == '-' {
			p.htmlReader.Discard(1)
			p.tokenBuilder.NewToken()
			return false, commentStartState, noError
		}

		return p.defaultMarkupDeclarationOpenStateParser()
	case 'D', 'd':
		peeked, err = p.htmlReader.Peek(peekDist)
		if err != nil {
			if len(peeked) < peekDist {
				return p.defaultMarkupDeclarationOpenStateParser()
			}
		}
		if bytes.EqualFold(peeked, doctype) {
			p.htmlReader.Discard(peekDist)
			return false, doctypeState, noError
		}
	case '[':
		peeked, err = p.htmlReader.Peek(peekDist)
		if err != nil {
			if len(peeked) < peekDist {
				return false, bogusCommentState, incorrectlyOpenedComment
			}
		}
		if bytes.Equal(cdata, peeked) {
			p.htmlReader.Discard(peekDist)
			if p.adjustedCurrentNode() && !p.inHTMLNamepsace() {
				return false, cdataSectionState, noError
			}
			p.tokenBuilder.NewToken()
			p.tokenBuilder.WriteData('[')
			p.tokenBuilder.WriteData('C')
			p.tokenBuilder.WriteData('D')
			p.tokenBuilder.WriteData('A')
			p.tokenBuilder.WriteData('T')
			p.tokenBuilder.WriteData('A')
			p.tokenBuilder.WriteData('[')
			return false, bogusCommentState, cdataInHTMLContent
		}
	default:
		return p.defaultMarkupDeclarationOpenStateParser()
	}

	return p.defaultMarkupDeclarationOpenStateParser()
}

func (p *HTMLTokenizer) commentStartStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, commentState, noError
	}
	switch r {
	case '-':
		return false, commentStartDashState, noError
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState, abruptClosingOfEmptyComment
	default:
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) commentStartDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CommentToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInComment
	}
	switch r {
	case '-':
		return false, commentEndState, noError
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState, abruptClosingOfEmptyComment
	default:
		p.tokenBuilder.WriteData('-')
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) commentStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CommentToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInComment
	}
	switch r {
	case '<':
		p.tokenBuilder.WriteData(r)
		return false, commentLessThanSignState, noError
	case '-':
		return false, commentEndDashState, noError
	case '\u0000':
		p.tokenBuilder.WriteData('\uFFFD')
		return false, commentState, unexpectedNullCharacter
	default:
		p.tokenBuilder.WriteData(r)
		return false, commentState, noError
	}
}
func (p *HTMLTokenizer) commentLessThanSignStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, commentState, noError
	}
	switch r {
	case '!':
		p.tokenBuilder.WriteData(r)
		return false, commentLessThanSignBangState, noError
	case '<':
		p.tokenBuilder.WriteData(r)
		return false, commentLessThanSignState, noError
	default:
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) commentLessThanSignBangStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, commentState, noError
	}
	switch r {
	case '-':
		return false, commentLessThanSignBangDashState, noError
	default:
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) commentLessThanSignBangDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, commentEndDashState, noError
	}
	switch r {
	case '-':
		return false, commentLessThanSignBangDashDashState, noError
	default:
		return true, commentEndDashState, noError
	}
}
func (p *HTMLTokenizer) commentLessThanSignBangDashDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, commentEndState, noError
	}
	switch r {
	case '>':
		return true, commentEndState, noError
	default:
		return true, commentEndState, nestedComment
	}
}
func (p *HTMLTokenizer) commentEndDashStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CommentToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInComment
	}
	switch r {
	case '-':
		return false, commentEndState, noError
	default:
		p.tokenBuilder.WriteData('-')
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) commentEndStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {

		p.emit(p.tokenBuilder.CommentToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInComment
	}
	switch r {
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState, noError
	case '!':
		return false, commentEndBangState, noError
	case '-':
		p.tokenBuilder.WriteData('-')
		return false, commentEndState, noError
	default:
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('-')
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) commentEndBangStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CommentToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInComment
	}
	switch r {
	case '-':
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('!')
		return false, commentEndDashState, noError
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState, incorrectlyClosedComment
	default:
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('!')
		return true, commentState, noError
	}
}
func (p *HTMLTokenizer) doctypeStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.NewToken()
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInComment
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeNameState, noError
	case '>':
		return true, beforeDoctypeNameState, noError
	default:
		return true, beforeDoctypeNameState, missingWhitespaceBeforeDoctypeName
	}
}
func (p *HTMLTokenizer) beforeDoctypeNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.NewToken()
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeNameState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.NewToken()
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState, noError
	case '\u0000':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.WriteName('\uFFFD')
		return false, doctypeNameState, unexpectedNullCharacter
	case '>':
		p.tokenBuilder.NewToken()
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, missingDoctypeName
	default:
		p.tokenBuilder.NewToken()
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState, noError
	}
}
func (p *HTMLTokenizer) doctypeNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterDoctypeNameState, noError
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState, noError
	case '\u0000':
		p.tokenBuilder.WriteName('\uFFFD')
		return false, doctypeNameState, unexpectedNullCharacter
	default:
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState, noError
	}
}
func (p *HTMLTokenizer) afterDoctypeNameStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterDoctypeNameState, noError
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, noError
	default:
		b, err := p.htmlReader.Peek(5)
		if err != nil {
			p.tokenBuilder.EnableForceQuirks()
			return true, bogusDoctypeState, invalidCharacterSequenceAfterDoctypeName
		}
		bs := bytes.Join([][]byte{{byte(r)}, b}, []byte{})
		if bytes.EqualFold(bs, []byte("PUBLIC")) {
			p.htmlReader.Discard(5)
			return false, afterDoctypePublicKeywordState, noError
		} else if bytes.EqualFold(bs, []byte("SYSTEM")) {
			p.htmlReader.Discard(5)
			return false, afterDoctypeSystemKeywordState, noError
		}
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, invalidCharacterSequenceAfterDoctypeName
	}
}
func (p *HTMLTokenizer) afterDoctypePublicKeywordStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypePublicIdentifierState, noError
	case '"':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierDoubleQuotedState, missingWhitespaceAfterDoctypePublicKeyword
	case '\'':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierSingleQuotedState, missingWhitespaceAfterDoctypePublicKeyword
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, missingDoctypePublicIdentifier
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, missingQuoteBeforeDoctypePublicIdentifier
	}
}
func (p *HTMLTokenizer) beforeDoctypePublicIdentifierStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypePublicIdentifierState, noError
	case '"':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierDoubleQuotedState, noError
	case '\'':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierSingleQuotedState, noError
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, missingDoctypePublicIdentifier
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, missingQuoteBeforeDoctypePublicIdentifier
	}
}
func (p *HTMLTokenizer) doctypePublicIdentifierDoubleQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '"':
		return false, afterDoctypePublicIdentifierState, noError
	case '\u0000':
		p.tokenBuilder.WritePublicIdentifier('\uFFFD')
		return false, doctypePublicIdentifierDoubleQuotedState, unexpectedNullCharacter
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, abruptDoctypePublicIdentifier
	default:
		p.tokenBuilder.WritePublicIdentifier(r)
		return false, doctypePublicIdentifierDoubleQuotedState, noError
	}
}
func (p *HTMLTokenizer) doctypePublicIdentifierSingleQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\'':
		return false, afterDoctypePublicIdentifierState, noError
	case '\u0000':
		p.tokenBuilder.WritePublicIdentifier('\uFFFD')
		return false, doctypePublicIdentifierSingleQuotedState, unexpectedNullCharacter
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, abruptDoctypePublicIdentifier
	default:
		p.tokenBuilder.WritePublicIdentifier(r)
		return false, doctypePublicIdentifierSingleQuotedState, noError
	}
}
func (p *HTMLTokenizer) afterDoctypePublicIdentifierStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, betweenDoctypePublicAndSystemIdentifiersState, noError
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, noError
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState, missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState, missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, missingQuoteBeforeDoctypeSystemIdentifier
	}
}
func (p *HTMLTokenizer) betweenDoctypePublicAndSystemIdentifiersStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, betweenDoctypePublicAndSystemIdentifiersState, noError
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, noError
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState, noError
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState, noError
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, missingQuoteBeforeDoctypeSystemIdentifier
	}
}

func (p *HTMLTokenizer) afterDoctypeSystemKeywordStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeSystemIdentifierState, noError
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState, missingWhitespaceAfterDoctypeSystemKeyword
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState, missingWhitespaceAfterDoctypeSystemKeyword
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, missingDoctypeSystemIdentifier
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, missingQuoteBeforeDoctypeSystemIdentifier
	}
}

func (p *HTMLTokenizer) beforeDoctypeSystemIdentifierStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeSystemIdentifierState, noError
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState, missingWhitespaceAfterDoctypeSystemKeyword
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState, missingWhitespaceAfterDoctypeSystemKeyword
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, missingDoctypeSystemIdentifier
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState, missingQuoteBeforeDoctypeSystemIdentifier
	}
}
func (p *HTMLTokenizer) doctypeSystemIdentifierDoubleQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '"':
		return false, afterDoctypeSystemIdentifierState, noError
	case '\u0000':
		p.tokenBuilder.WriteSystemIdentifier('\uFFFD')
		return false, doctypeSystemIdentifierDoubleQuotedState, unexpectedNullCharacter
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, abruptDoctypeSystemIdentifier
	default:
		p.tokenBuilder.WriteSystemIdentifier(r)
		return false, doctypeSystemIdentifierDoubleQuotedState, noError
	}
}
func (p *HTMLTokenizer) doctypeSystemIdentifierSingleQuotedStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\'':
		return false, afterDoctypeSystemIdentifierState, noError
	case '\u0000':
		p.tokenBuilder.WriteSystemIdentifier('\uFFFD')
		return false, doctypeSystemIdentifierSingleQuotedState, unexpectedNullCharacter
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, abruptDoctypeSystemIdentifier
	default:
		p.tokenBuilder.WriteSystemIdentifier(r)
		return false, doctypeSystemIdentifierSingleQuotedState, noError
	}
}
func (p *HTMLTokenizer) afterDoctypeSystemIdentifierStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInDoctype
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterDoctypeSystemIdentifierState, noError
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, noError
	default:
		return true, bogusDoctypeState, unexpectedCharacterAfterDoctypeSystemIdentifier
	}
}
func (p *HTMLTokenizer) bogusDoctypeStateParser(r rune) (bool, tokenizerState, parseError) {

	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.DocTypeToken())
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, noError
	}
	switch r {
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState, noError
	case '\u0000':
		return false, bogusDoctypeState, unexpectedNullCharacter
	default:
		return false, bogusDoctypeState, noError
	}
}
func (p *HTMLTokenizer) cdataSectionStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState, eofInCdata
	}
	switch r {
	case ']':
		return false, cdataSectionBracketState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, cdataSectionState, noError
	}
}
func (p *HTMLTokenizer) cdataSectionBracketStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState, noError
	}
	switch r {
	case ']':
		return false, cdataSectionEndState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState, noError
	}
}
func (p *HTMLTokenizer) cdataSectionEndStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.emit(p.tokenBuilder.CharacterToken(']'))
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState, noError
	}
	switch r {
	case ']':
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return false, cdataSectionEndState, noError
	case '>':
		return false, dataState, noError
	default:
		p.emit(p.tokenBuilder.CharacterToken(']'))
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState, noError
	}
}

func (p *HTMLTokenizer) characterReferenceStateParser(r rune) (bool, tokenizerState, parseError) {
	p.tokenBuilder.ResetTempBuffer()
	p.tokenBuilder.WriteTempBuffer('&')

	if p.isEndOfFile() {
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState, noError
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true, namedCharacterReferenceState, noError
	case '#':
		p.tokenBuilder.WriteTempBuffer(r)
		return false, numericCharacterReferenceState, noError
	default:
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState, noError
	}
}

func (p *HTMLTokenizer) anyFilteredtable(filteredTable map[string][]rune, match string) (bool, string) {
	smallest := ""
	// one byte at a time, remove things that don't match
	for k := range filteredTable {
		if !strings.HasPrefix(k, match) {
			delete(filteredTable, k)
			continue
		}

		// if there is an exact match, we need to add these characters to the
		// temp buffer
		if k == match {
			smallest = k
			p.tokenBuilder.ResetTempBuffer()
			// don't erase the & character from character reference state though
			p.tokenBuilder.WriteTempBuffer('&')
			for _, r := range smallest {
				p.tokenBuilder.WriteTempBuffer(r)
			}
		}
	}

	return len(filteredTable) > 0, smallest
}

func (p *HTMLTokenizer) namedCharacterReferenceStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.flushCodePointsAsCharacterReference()
		return false, ambiguousAmpersandState, noError
	}

	filteredTable := make(map[string][]rune, len(charRefTable))
	for k, v := range charRefTable {
		filteredTable[k] = v
	}

	consumed := []byte{byte(r)}
	hasMatches, smallest := p.anyFilteredtable(filteredTable, string(consumed))

	var (
		b, sep, next []byte
		err          error
		newSmallest  string
		firstDiscard bool = true
	)
	match := bytes.Join([][]byte{consumed, b}, sep)

	// consume one character until there are no more matches in the filtered table
	for i := 1; hasMatches; i++ {
		b, err = p.htmlReader.Peek(i)
		// if we can't peek any more bytes, leave the loop and use the currently
		// calculated smallest match as the match.
		if err != nil {
			break
		}

		// concatenate the read bytes to the consumed
		match = bytes.Join([][]byte{consumed, b}, sep)
		hasMatches, newSmallest = p.anyFilteredtable(filteredTable, string(match))

		// if we found a larger, better match in the table, update the reader and
		// the loop counter
		if len(newSmallest) > len(smallest) {
			i = 0
			diff := len(newSmallest) - len(smallest)

			// because we already dicarded the first rune in the main loop
			// we don't need discard it again here.
			if firstDiscard {
				_, err = p.htmlReader.Discard(diff - 1)
				firstDiscard = false
			} else {
				_, err = p.htmlReader.Discard(diff)
			}
			if err != nil {
				break
			}

			smallest = newSmallest
			consumed = match
		}
	}

	// there wasn't a match in the table, we haven't added to the temp buffer yet
	// so when we try to flush the consumed code points, nothing will be there.
	// add the consumed character to the temp buffer in this case.
	if smallest == "" {
		for i, mb := range match {
			if i != 0 {
				p.htmlReader.Discard(1)
			}
			p.tokenBuilder.WriteTempBuffer(rune(mb))
		}
		p.flushCodePointsAsCharacterReference()
		return false, ambiguousAmpersandState, noError
	}

	endsInSemiColon := bytes.HasSuffix(consumed, []byte{';'})
	if p.wasConsumedByAttribute() && !endsInSemiColon {
		next, err = p.htmlReader.Peek(1)
		if err == nil {
			switch next[0] {
			case '=', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				for _, r := range next {
					p.tokenBuilder.WriteTempBuffer(rune(r))
					p.htmlReader.Discard(1)
				}
				p.flushCodePointsAsCharacterReference()
				return false, p.returnState, noError
			}
		}
	}

	errorType := noError
	if !endsInSemiColon {
		errorType = missingSemiColonAfterCharacterReference
	}
	p.tokenBuilder.ResetTempBuffer()
	for _, r := range charRefTable[string(consumed)] {
		p.tokenBuilder.WriteTempBuffer(r)
	}
	p.flushCodePointsAsCharacterReference()

	return false, p.returnState, errorType

}
func (p *HTMLTokenizer) ambiguousAmpersandStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, p.returnState, noError
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if p.wasConsumedByAttribute() {
			p.tokenBuilder.WriteAttributeValue(r)
		} else {
			p.emit(p.tokenBuilder.CharacterToken(r))
		}
		return false, ambiguousAmpersandState, noError
	case ';':
		return true, p.returnState, unknownNamedCharacterReference
	default:
		return true, p.returnState, noError
	}
}
func (p *HTMLTokenizer) numericCharacterReferenceStateParser(r rune) (bool, tokenizerState, parseError) {
	p.tokenBuilder.SetCharRef(0)
	if p.isEndOfFile() {
		return true, decimalCharacterReferenceStartState, noError
	}
	switch r {
	case 'x', 'X':
		p.tokenBuilder.WriteTempBuffer(r)
		return false, hexadecimalCharacterReferenceStartState, noError
	default:
		return true, decimalCharacterReferenceStartState, noError
	}
}
func (p *HTMLTokenizer) hexadecimalCharacterReferenceStartStateParser(r rune) (bool, tokenizerState, parseError) {

	if p.isEndOfFile() {
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState, absenceOfDigitsInNumericCharacterReference
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'a', 'b', 'c', 'd', 'e', 'f':
		return true, hexadecimalCharacterReferenceState, noError
	default:
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState, absenceOfDigitsInNumericCharacterReference
	}
}
func (p *HTMLTokenizer) decimalCharacterReferenceStartStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState, absenceOfDigitsInNumericCharacterReference
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true, decimalCharacterReferenceState, noError
	default:
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState, absenceOfDigitsInNumericCharacterReference
	}
}
func (p *HTMLTokenizer) hexadecimalCharacterReferenceStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, numericCharacterReferenceEndState, missingSemiColonAfterCharacterReference
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.tokenBuilder.MultByCharRef(16)
		p.tokenBuilder.AddToCharRef(int(r - 0x30))
		return false, hexadecimalCharacterReferenceState, noError
	case 'A', 'B', 'C', 'D', 'E', 'F':
		p.tokenBuilder.MultByCharRef(16)
		p.tokenBuilder.AddToCharRef(int(r - 0x37))
		return false, hexadecimalCharacterReferenceState, noError
	case 'a', 'b', 'c', 'd', 'e', 'f':
		p.tokenBuilder.MultByCharRef(16)
		p.tokenBuilder.AddToCharRef(int(r - 0x57))
		return false, hexadecimalCharacterReferenceState, noError
	case ';':
		return false, numericCharacterReferenceEndState, noError
	default:
		return true, numericCharacterReferenceEndState, missingSemiColonAfterCharacterReference
	}
}
func (p *HTMLTokenizer) decimalCharacterReferenceStateParser(r rune) (bool, tokenizerState, parseError) {
	if p.isEndOfFile() {
		return true, numericCharacterReferenceEndState, missingSemiColonAfterCharacterReference
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.tokenBuilder.MultByCharRef(10)
		p.tokenBuilder.AddToCharRef(int(r - 0x30))
		return false, decimalCharacterReferenceState, noError
	case ';':
		return false, numericCharacterReferenceEndState, noError
	default:
		return true, numericCharacterReferenceEndState, missingSemiColonAfterCharacterReference
	}
}

var numericCharacterReferenceEndStateTable map[int]rune = map[int]rune{
	0x80: 0x20AC,
	0x82: 0x201A,
	0x83: 0x0192,
	0x84: 0x201E,
	0x85: 0x2026,
	0x86: 0x2020,
	0x87: 0x2021,
	0x88: 0x02C6,
	0x89: 0x2030,
	0x8A: 0x0160,
	0x8B: 0x2039,
	0x8C: 0x0152,
	0x8E: 0x017D,
	0x91: 0x2018,
	0x92: 0x2019,
	0x93: 0x201C,
	0x94: 0x201D,
	0x95: 0x2022,
	0x96: 0x2013,
	0x97: 0x2014,
	0x98: 0x02DC,
	0x99: 0x2122,
	0x9A: 0x0161,
	0x9B: 0x203A,
	0x9C: 0x0153,
	0x9E: 0x017E,
	0x9F: 0x0178,
}

func (p *HTMLTokenizer) numericCharacterReferenceEndStateParser(r rune) (bool, tokenizerState, parseError) {
	// this is the only state that isn't suppose to consume something off the bat.
	p.htmlReader.UnreadRune()
	err := noError
	if p.tokenBuilder.Cmp(0) == 0 {
		err = nullCharacterReference
		p.tokenBuilder.SetCharRef(0xFFFD)
	} else if p.tokenBuilder.Cmp(0x10FFFF) == 1 {
		err = characterReferenceOutsideUnicodeRange
		p.tokenBuilder.SetCharRef(0xFFFD)
	} else if isSurrogate(p.tokenBuilder.GetCharRef()) {
		err = surrogateCharacterReference
		p.tokenBuilder.SetCharRef(0xFFFD)
	} else if isNonCharacter(p.tokenBuilder.GetCharRef()) {
		err = noncharacterCharacterReference
	} else if p.tokenBuilder.Cmp(0x0D) == 0 ||
		(isControl(p.tokenBuilder.GetCharRef())) && !isASCIIWhitespace(p.tokenBuilder.GetCharRef()) {
		err = controlCharacterReference
		tableNum, ok := numericCharacterReferenceEndStateTable[p.tokenBuilder.GetCharRef()]
		if ok {
			p.tokenBuilder.SetCharRef(int(tableNum))
		}
	}

	p.tokenBuilder.ResetTempBuffer()
	p.tokenBuilder.WriteTempBuffer(rune(p.tokenBuilder.GetCharRef()))
	p.flushCodePointsAsCharacterReference()
	return false, p.returnState, err
}

func (p *HTMLTokenizer) emitCurrentTag() {
	if p.tokenBuilder.curTagType == startTag {
		p.emit(p.tokenBuilder.StartTagToken())
	} else {
		p.emit(p.tokenBuilder.EndTagToken())
	}
}

// a stateHandler is a func that takes in a rune
// and returns the next state to transition to.
type parserStateHandler func(in rune) (bool, tokenizerState, parseError)

//go:generate stringer -type=parseError
type parseError uint

const (
	noError parseError = iota
	abruptClosingOfEmptyComment
	abruptDoctypePublicIdentifier
	abruptDoctypeSystemIdentifier
	absenceOfDigitsInNumericCharacterReference
	cdataInHTMLContent
	characterReferenceOutsideUnicodeRange
	controlCharacterInInputSteam
	controlCharacterReference
	endTagWithAttributes
	duplicateAttribute
	endTagWithTrailingSolidus
	eofBeforeTagName
	eofInCdata
	eofInComment
	eofInDoctype
	eofInScriptHTMLCommentLikeText
	eofInTag
	incorrectlyClosedComment
	incorrectlyOpenedComment
	invalidCharacterSequenceAfterDoctypeName
	invalidFirstCharacterOfTagName
	missingAttributeValue
	missingDoctypeName
	missingDoctypePublicIdentifier
	missingDoctypeSystemIdentifier
	missingEndTagName
	missingQuoteBeforeDoctypePublicIdentifier
	missingQuoteBeforeDoctypeSystemIdentifier
	missingSemiColonAfterCharacterReference
	missingWhitespaceAfterDoctypePublicKeyword
	missingWhitespaceAfterDoctypeSystemKeyword
	missingWhitespaceBeforeDoctypeName
	missingWhitespaceBetweenAttributes
	missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers
	nestedComment
	noncharacterCharacterReference
	noncharacterInInputStream
	nonVoidHTMLElementStartTagWithTrailingSolidus
	nullCharacterReference
	surrogateCharacterReference
	surrogateInInputStream
	unexpectedCharacterAfterDoctypeSystemIdentifier
	unexpectedCharacterInAttributeName
	unexpectedCharacterInUnquotedAttributeValue
	unexpectedEqualsSignBeforeAttributeName
	unexpectedNullCharacter
	unexpectedQuestionMakrInsteadofTagName
	unexpectedSolidusInTag
	unknownNamedCharacterReference
	generalParseError
)

//go:generate stringer -type=tokenizerState
type tokenizerState uint

const (
	dataState tokenizerState = iota
	rcDataState
	rawTextState
	scriptDataState
	plaintextState
	tagOpenState
	endTagOpenState
	tagNameState
	rcDataLessThanSignState
	rcDataEndTagOpenState
	rcDataEndTagNameState
	rawTextLessThanSignState
	rawTextEndTagOpenState
	rawTextEndTagNameState
	scriptDataLessThanSignState
	scriptDataEndTagOpenState
	scriptDataEndTagNameState
	scriptDataEscapeStartState
	scriptDataEscapeStartDashState
	scriptDataEscapedState
	scriptDataEscapedDashState
	scriptDataEscapedDashDashState
	scriptDataEscapedLessThanSignState
	scriptDataEscapedEndTagOpenState
	scriptDataEscapedEndTagNameState
	scriptDataDoubleEscapeStartState
	scriptDataDoubleEscapedState
	scriptDataDoubleEscapedDashState
	scriptDataDoubleEscapedDashDashState
	scriptDataDoubleEscapedLessThanSignState
	scriptDataDoubleEscapeEndState
	beforeAttributeNameState
	attributeNameState
	afterAttributeNameState
	beforeAttributeValueState
	attributeValueDoubleQuotedState
	attributeValueSingleQuotedState
	attributeValueUnquotedState
	afterAttributeValueQuotedState
	selfClosingStartTagState
	bogusCommentState
	markupDeclarationOpenState
	commentStartState
	commentStartDashState
	commentState
	commentLessThanSignState
	commentLessThanSignBangState
	commentLessThanSignBangDashState
	commentLessThanSignBangDashDashState
	commentEndDashState
	commentEndState
	commentEndBangState
	doctypeState
	beforeDoctypeNameState
	doctypeNameState
	afterDoctypeNameState
	afterDoctypePublicKeywordState
	beforeDoctypePublicIdentifierState
	doctypePublicIdentifierDoubleQuotedState
	doctypePublicIdentifierSingleQuotedState
	afterDoctypePublicIdentifierState
	betweenDoctypePublicAndSystemIdentifiersState
	afterDoctypeSystemKeywordState
	beforeDoctypeSystemIdentifierState
	doctypeSystemIdentifierDoubleQuotedState
	doctypeSystemIdentifierSingleQuotedState
	afterDoctypeSystemIdentifierState
	bogusDoctypeState
	cdataSectionState
	cdataSectionBracketState
	cdataSectionEndState
	characterReferenceState
	namedCharacterReferenceState
	ambiguousAmpersandState
	numericCharacterReferenceState
	hexadecimalCharacterReferenceStartState
	decimalCharacterReferenceStartState
	hexadecimalCharacterReferenceState
	decimalCharacterReferenceState
	numericCharacterReferenceEndState
)

// Tokenize the current HTMLParser's htmlString attribute and emits
// tokens to the token channel to be consumed by the TreeConstruction
// process or another process that wants to use the tokens.
// The choice to parse this as a string is that it makes it easier to iterate over
// rune values with a for/range (go by default pulls utf-8 characters and produces runes
// on each iteration, so we may need to switch this if we want to support other uncoding types)
func (p *HTMLTokenizer) Tokenize() {
	defer func() {
		close(p.tokenChannel)
		p.wg.Done()
	}()
	lastState := p.tokenizeUntilEOF(dataState)
	p.tokenizeEOF(lastState)
}

func logError(e parseError) {

}

func (p *HTMLTokenizer) normalizeNewlines(r rune) rune {
	if r == '\u000D' {
		b, err := p.htmlReader.Peek(1)
		if err != nil {
			return '\u000A'
		}
		if len(b) > 0 && b[0] == '\u000A' {
			p.htmlReader.Discard(1)
		}

		return '\u000A'
	}

	return r
}

// tokenizeUntilEOF is a helper function that performs the tokenization provided an initial
// state. It's convinent when we want to run tests that start in the middle of the state
// machine. It also doesn't process the EOF states, which does a lot of cleanup and allows
// us to test how certain characters affect the tokenbuilder.
func (p *HTMLTokenizer) tokenizeUntilEOF(nextState tokenizerState) tokenizerState {
	defer p.wg.Done()
	var (
		reconsume bool
		err       error
		parseErr  parseError
		r         rune
	)
	nextStateHandler := p.mappings[nextState]
	for {
		r, _, err = p.htmlReader.ReadRune()
		if err != nil {
			if err == io.EOF {
				p.eof = true
				break
			}
		}

		r = p.normalizeNewlines(r)

		reconsume, nextState, parseErr = nextStateHandler(r)
		if p.config[debug] == 1 {
			logError(parseErr)
		}
		nextStateHandler = p.mappings[nextState]
		// if the previous state says to reconsume,
		// use the same rune
		for reconsume {
			reconsume, nextState, parseErr = nextStateHandler(r)
			if p.config[debug] == 1 {
				logError(parseErr)
			}
			nextStateHandler = p.mappings[nextState]
		}
	}

	return nextState
}

// tokenizeEOF takes the last state before an EOF was found and runs through its state machine calling
// any reconsume or EOF handlers. Some handlers don't have a EOF handler, so that is why we have to
// loop through any reconsumes until we get one.
func (p *HTMLTokenizer) tokenizeEOF(nextState tokenizerState) {
	var (
		nextStateHandler parserStateHandler
		reconsume        bool = true
		parseErr         parseError
		r                rune
	)

	for reconsume {
		// last state
		nextStateHandler = p.mappings[nextState]
		reconsume, nextState, parseErr = nextStateHandler(r)
		if p.config[debug] == 1 {
			logError(parseErr)
		}
	}
}
