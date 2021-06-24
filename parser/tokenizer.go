package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/heathj/gobrowse/parser/spec"
)

// HTMLTokenizer holds state for the various state of the tokenizer.
type HTMLTokenizer struct {
	done                      bool
	returnState, currentState tokenizerState
	inputStream               *bufio.Reader
	adjustedCurrentNode       *spec.Node
	emittedTokens             []Token
	tokenBuilder              *TokenBuilder
	lastEmittedStartTagName   string
}

// NewHTMLTokenizer creates an HTML parser that can be used to process
// an HTML string.
func NewHTMLTokenizer(io io.Reader) *HTMLTokenizer {
	return &HTMLTokenizer{
		emittedTokens: []Token{},
		inputStream:   bufio.NewReader(io),
		tokenBuilder:  MakeTokenBuilder(),
	}
}

func (p *HTMLTokenizer) stateToParser(state tokenizerState) parserStateHandler {
	switch state {
	case dataState:
		return p.dataStateParser
	case rcDataState:
		return p.rcDataStateParser
	case rawTextState:
		return p.rawTextStateParser
	case scriptDataState:
		return p.scriptDataStateParser
	case plaintextState:
		return p.plaintextStateParser
	case tagOpenState:
		return p.tagOpenStateParser
	case endTagOpenState:
		return p.endTagOpenStateParser
	case tagNameState:
		return p.tagNameStateParser
	case rcDataLessThanSignState:
		return p.rcDataLessThanSignStateParser
	case rcDataEndTagOpenState:
		return p.rcDataEndTagOpenStateParser
	case rcDataEndTagNameState:
		return p.rcDataEndTagNameStateParser
	case rawTextLessThanSignState:
		return p.rawTextLessThanSignStateParser
	case rawTextEndTagOpenState:
		return p.rawTextEndTagOpenStateParser
	case rawTextEndTagNameState:
		return p.rawTextEndTagNameStateParser
	case scriptDataLessThanSignState:
		return p.scriptDataLessThanSignStateParser
	case scriptDataEndTagOpenState:
		return p.scriptDataEndTagOpenStateParser
	case scriptDataEndTagNameState:
		return p.scriptDataEndTagNameStateParser
	case scriptDataEscapeStartState:
		return p.scriptDataEscapeStartStateParser
	case scriptDataEscapeStartDashState:
		return p.scriptDataEscapeStartDashStateParser
	case scriptDataEscapedState:
		return p.scriptDataEscapedStateParser
	case scriptDataEscapedDashState:
		return p.scriptDataEscapedDashStateParser
	case scriptDataEscapedDashDashState:
		return p.scriptDataEscapedDashDashStateParser
	case scriptDataEscapedLessThanSignState:
		return p.scriptDataEscapedLessThanSignStateParser
	case scriptDataEscapedEndTagOpenState:
		return p.scriptDataEscapedEndTagOpenStateParser
	case scriptDataEscapedEndTagNameState:
		return p.scriptDataEscapedEndTagNameStateParser
	case scriptDataDoubleEscapeStartState:
		return p.scriptDataDoubleEscapeStartStateParser
	case scriptDataDoubleEscapedState:
		return p.scriptDataDoubleEscapedStateParser
	case scriptDataDoubleEscapedDashState:
		return p.scriptDataDoubleEscapedDashStateParser
	case scriptDataDoubleEscapedDashDashState:
		return p.scriptDataDoubleEscapedDashDashStateParser
	case scriptDataDoubleEscapedLessThanSignState:
		return p.scriptDataDoubleEscapedLessThanSignStateParser
	case scriptDataDoubleEscapeEndState:
		return p.scriptDataDoubleEscapeEndStateParser
	case beforeAttributeNameState:
		return p.beforeAttributeNameStateParser
	case attributeNameState:
		return p.attributeNameStateParser
	case afterAttributeNameState:
		return p.afterAttributeNameStateParser
	case beforeAttributeValueState:
		return p.beforeAttributeValueStateParser
	case attributeValueDoubleQuotedState:
		return p.attributeValueDoubleQuotedStateParser
	case attributeValueSingleQuotedState:
		return p.attributeValueSingleQuotedStateParser
	case attributeValueUnquotedState:
		return p.attributeValueUnquotedStateParser
	case afterAttributeValueQuotedState:
		return p.afterAttributeValueQuotedStateParser
	case selfClosingStartTagState:
		return p.selfClosingStartTagStateParser
	case bogusCommentState:
		return p.bogusCommentStateParser
	case markupDeclarationOpenState:
		return p.markupDeclarationOpenStateParser
	case commentStartState:
		return p.commentStartStateParser
	case commentStartDashState:
		return p.commentStartDashStateParser
	case commentState:
		return p.commentStateParser
	case commentLessThanSignState:
		return p.commentLessThanSignStateParser
	case commentLessThanSignBangState:
		return p.commentLessThanSignBangStateParser
	case commentLessThanSignBangDashState:
		return p.commentLessThanSignBangDashStateParser
	case commentLessThanSignBangDashDashState:
		return p.commentLessThanSignBangDashDashStateParser
	case commentEndDashState:
		return p.commentEndDashStateParser
	case commentEndState:
		return p.commentEndStateParser
	case commentEndBangState:
		return p.commentEndBangStateParser
	case doctypeState:
		return p.doctypeStateParser
	case beforeDoctypeNameState:
		return p.beforeDoctypeNameStateParser
	case doctypeNameState:
		return p.doctypeNameStateParser
	case afterDoctypeNameState:
		return p.afterDoctypeNameStateParser
	case afterDoctypePublicKeywordState:
		return p.afterDoctypePublicKeywordStateParser
	case beforeDoctypePublicIdentifierState:
		return p.beforeDoctypePublicIdentifierStateParser
	case doctypePublicIdentifierDoubleQuotedState:
		return p.doctypePublicIdentifierDoubleQuotedStateParser
	case doctypePublicIdentifierSingleQuotedState:
		return p.doctypePublicIdentifierSingleQuotedStateParser
	case afterDoctypePublicIdentifierState:
		return p.afterDoctypePublicIdentifierStateParser
	case betweenDoctypePublicAndSystemIdentifiersState:
		return p.betweenDoctypePublicAndSystemIdentifiersStateParser
	case afterDoctypeSystemKeywordState:
		return p.afterDoctypeSystemKeywordStateParser
	case beforeDoctypeSystemIdentifierState:
		return p.beforeDoctypeSystemIdentifierStateParser
	case doctypeSystemIdentifierDoubleQuotedState:
		return p.doctypeSystemIdentifierDoubleQuotedStateParser
	case doctypeSystemIdentifierSingleQuotedState:
		return p.doctypeSystemIdentifierSingleQuotedStateParser
	case afterDoctypeSystemIdentifierState:
		return p.afterDoctypeSystemIdentifierStateParser
	case bogusDoctypeState:
		return p.bogusDoctypeStateParser
	case cdataSectionState:
		return p.cdataSectionStateParser
	case cdataSectionBracketState:
		return p.cdataSectionBracketStateParser
	case cdataSectionEndState:
		return p.cdataSectionEndStateParser
	case characterReferenceState:
		return p.characterReferenceStateParser
	case namedCharacterReferenceState:
		return p.namedCharacterReferenceStateParser
	case ambiguousAmpersandState:
		return p.ambiguousAmpersandStateParser
	case numericCharacterReferenceState:
		return p.numericCharacterReferenceStateParser
	case hexadecimalCharacterReferenceStartState:
		return p.hexadecimalCharacterReferenceStartStateParser
	case decimalCharacterReferenceStartState:
		return p.decimalCharacterReferenceStartStateParser
	case hexadecimalCharacterReferenceState:
		return p.hexadecimalCharacterReferenceStateParser
	case decimalCharacterReferenceState:
		return p.decimalCharacterReferenceStateParser
	case numericCharacterReferenceEndState:
		return p.numericCharacterReferenceEndStateParser
	}

	return nil
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

func wasConsumedByAttribute(returnState tokenizerState) bool {
	switch returnState {
	case attributeValueDoubleQuotedState, attributeValueSingleQuotedState, attributeValueUnquotedState:
		return true
	}
	return false
}

func (p *HTMLTokenizer) flushCodePointsAsCharacterReference() {
	if wasConsumedByAttribute(p.returnState) {
		for _, v := range p.tokenBuilder.TempBuffer() {
			p.tokenBuilder.WriteAttributeValue(v)
		}
	} else {
		p.emit(p.tokenBuilder.TempBufferCharTokens()...)
	}
}

func (p *HTMLTokenizer) isApprEndTagToken() bool {
	return p.lastEmittedStartTagName == p.tokenBuilder.name.String()
}

func (p *HTMLTokenizer) emit(tokens ...Token) {
	for _, token := range tokens {
		if token.TokenType == endTagToken {
			if len(token.Attributes) > 0 {
				token.Attributes = make(map[string]*spec.Attr)
			}
			if token.SelfClosing {
				token.SelfClosing = false
			}
		} else if token.TokenType == startTagToken {
			p.lastEmittedStartTagName = token.TagName
		}

		p.emittedTokens = append(p.emittedTokens, token)
	}
}

func (p *HTMLTokenizer) dataStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '&':
		p.returnState = dataState
		return false, characterReferenceState
	case '<':
		return false, tagOpenState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, dataState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, dataState
	}
}

func (p *HTMLTokenizer) rcDataStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '&':
		p.returnState = rcDataState
		return false, characterReferenceState
	case '<':
		return false, rcDataLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, rcDataState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, rcDataState
	}
}
func (p *HTMLTokenizer) rawTextStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '<':
		return false, rawTextLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, rawTextState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, rawTextState
	}
}
func (p *HTMLTokenizer) scriptDataStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '<':
		return false, scriptDataLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataState
	}
}

func (p *HTMLTokenizer) plaintextStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, plaintextState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, plaintextState
	}
}

func (p *HTMLTokenizer) tagOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '!':
		return false, markupDeclarationOpenState
	case '/':
		return false, endTagOpenState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		p.tokenBuilder.curTagType = startTag
		return true, tagNameState
	case '?':
		p.tokenBuilder.Reset()
		return true, bogusCommentState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, dataState
	}
}

func (p *HTMLTokenizer) endTagOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		p.tokenBuilder.curTagType = endTag
		return true, tagNameState
	case '>':

		return false, dataState
	default:
		p.tokenBuilder.Reset()
		return true, bogusCommentState
	}
}

func (p *HTMLTokenizer) tagNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020': // tab, line feed, form feed, space
		return false, beforeAttributeNameState
	case '/':
		return false, selfClosingStartTagState
	case '>':
		return false, p.emitCurrentTag()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, tagNameState
	case '\u0000': // null
		p.tokenBuilder.WriteName('\uFFFD')
		return false, tagNameState
	default:
		p.tokenBuilder.WriteName(r)
		return false, tagNameState
	}
}

func (p *HTMLTokenizer) rcDataLessThanSignStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rcDataState
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, rcDataEndTagOpenState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rcDataState
	}
}

func (p *HTMLTokenizer) defaultRcDataEndTagOpenStateParser() (bool, tokenizerState) {
	p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
	return true, rcDataState
}
func (p *HTMLTokenizer) rcDataEndTagOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return p.defaultRcDataEndTagOpenStateParser()
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		p.tokenBuilder.curTagType = endTag
		return true, rcDataEndTagNameState
	default:
		return p.defaultRcDataEndTagOpenStateParser()
	}
}

func (p *HTMLTokenizer) defaultRcDataEndTagNameStateCase() (bool, tokenizerState) {
	p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
	p.emit(p.tokenBuilder.TempBufferCharTokens()...)
	return true, rcDataState
}
func (p *HTMLTokenizer) rcDataEndTagNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return p.defaultRcDataEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState
		}
		return p.defaultRcDataEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState
		}
		return p.defaultRcDataEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			return false, p.emitCurrentTag()
		}
		return p.defaultRcDataEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, rcDataEndTagNameState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, rcDataEndTagNameState
	default:
		return p.defaultRcDataEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) rawTextLessThanSignStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rawTextState
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, rawTextEndTagOpenState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, rawTextState
	}
}

func (p *HTMLTokenizer) defaultRawTextEndTagOpenStateParser() (bool, tokenizerState) {
	p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
	return true, rawTextState
}

func (p *HTMLTokenizer) rawTextEndTagOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return p.defaultRawTextEndTagOpenStateParser()
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		p.tokenBuilder.curTagType = endTag
		return true, rawTextEndTagNameState
	default:
		return p.defaultRawTextEndTagOpenStateParser()
	}
}

func (p *HTMLTokenizer) defaultRawTextEndTagNameStateCase() (bool, tokenizerState) {
	p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
	p.emit(p.tokenBuilder.TempBufferCharTokens()...)
	return true, rawTextState
}
func (p *HTMLTokenizer) rawTextEndTagNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return p.defaultRawTextEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState
		}
		return p.defaultRawTextEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState
		}
		return p.defaultRawTextEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			return false, p.emitCurrentTag()
		}
		return p.defaultRawTextEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, rawTextEndTagNameState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, rawTextEndTagNameState
	default:
		return p.defaultRawTextEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) scriptDataLessThanSignStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataState
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, scriptDataEndTagOpenState
	case '!':
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('!'))
		return false, scriptDataEscapeStartState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataState
	}
}

func (p *HTMLTokenizer) scriptDataEndTagOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataState
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		p.tokenBuilder.curTagType = endTag
		return true, scriptDataEndTagNameState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataState
	}
}

func (p *HTMLTokenizer) defaultScriptDataEndTagNameStateCase() (bool, tokenizerState) {
	p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
	p.emit(p.tokenBuilder.TempBufferCharTokens()...)
	return true, scriptDataState
}
func (p *HTMLTokenizer) scriptDataEndTagNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return p.defaultScriptDataEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState
		}
		return p.defaultScriptDataEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState
		}
		return p.defaultScriptDataEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			return false, p.emitCurrentTag()
		}
		return p.defaultScriptDataEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEndTagNameState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEndTagNameState
	default:
		return p.defaultScriptDataEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) scriptDataEscapeStartStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, scriptDataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapeStartDashState
	default:
		return true, scriptDataState
	}
}
func (p *HTMLTokenizer) scriptDataEscapeStartDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, scriptDataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashDashState
	default:
		return true, scriptDataState
	}
}
func (p *HTMLTokenizer) scriptDataEscapedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashState
	case '<':
		return false, scriptDataEscapedLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataEscapedState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataEscapedDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashDashState
	case '<':
		return false, scriptDataEscapedLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataEscapedState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataEscapedDashDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataEscapedDashDashState
	case '<':
		return false, scriptDataEscapedLessThanSignState
	case '>':
		p.emit(p.tokenBuilder.CharacterToken('>'))
		return false, scriptDataState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataEscapedState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataEscapedLessThanSignStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataEscapedState
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		return false, scriptDataEscapedEndTagOpenState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.ResetTempBuffer()
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataDoubleEscapeStartState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return true, scriptDataEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataEscapedEndTagOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataEscapedState
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		p.tokenBuilder.curTagType = endTag
		return true, scriptDataEscapedEndTagNameState
	default:
		p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
		return true, scriptDataEscapedState
	}
}

func (p *HTMLTokenizer) defaultScriptDataEscapedEndTagNameStateCase() (bool, tokenizerState) {
	p.emit(p.tokenBuilder.CharacterToken('<'), p.tokenBuilder.CharacterToken('/'))
	p.emit(p.tokenBuilder.TempBufferCharTokens()...)
	return true, scriptDataEscapedState
}
func (p *HTMLTokenizer) scriptDataEscapedEndTagNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		if p.isApprEndTagToken() {
			return false, beforeAttributeNameState
		}
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	case '/':
		if p.isApprEndTagToken() {
			return false, selfClosingStartTagState
		}
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	case '>':
		if p.isApprEndTagToken() {
			return false, p.emitCurrentTag()
		}
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteTempBuffer(r)
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEscapedEndTagNameState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.tokenBuilder.WriteTempBuffer(r)
		p.tokenBuilder.WriteName(r)
		return false, scriptDataEscapedEndTagNameState
	default:
		return p.defaultScriptDataEscapedEndTagNameStateCase()
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapeStartStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, scriptDataEscapedState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020', '/', '>':
		p.emit(p.tokenBuilder.CharacterToken(r))
		if p.tokenBuilder.TempBuffer() == "script" {
			return false, scriptDataDoubleEscapedState
		}
		return false, scriptDataEscapedState
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		r += 0x20
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeStartState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeStartState
	default:
		return true, scriptDataEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataDoubleEscapedDashState
	case '<':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return false, scriptDataDoubleEscapedLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataDoubleEscapedState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataDoubleEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataDoubleEscapedDashDashState
	case '<':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return false, scriptDataDoubleEscapedLessThanSignState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataDoubleEscapedState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataDoubleEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedDashDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.emit(p.tokenBuilder.CharacterToken('-'))
		return false, scriptDataDoubleEscapedDashDashState
	case '<':
		p.emit(p.tokenBuilder.CharacterToken('<'))
		return false, scriptDataDoubleEscapedLessThanSignState
	case '>':
		p.emit(p.tokenBuilder.CharacterToken('>'))
		return false, scriptDataState
	case '\u0000':
		p.emit(p.tokenBuilder.CharacterToken('\uFFFD'))
		return false, scriptDataDoubleEscapedState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, scriptDataDoubleEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapedLessThanSignStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, scriptDataDoubleEscapedState
	}
	switch r {
	case '/':
		p.tokenBuilder.ResetTempBuffer()
		p.emit(p.tokenBuilder.CharacterToken('/'))
		return false, scriptDataDoubleEscapeEndState
	default:
		return true, scriptDataDoubleEscapedState
	}
}
func (p *HTMLTokenizer) scriptDataDoubleEscapeEndStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, scriptDataDoubleEscapedState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020', '/', '>':
		p.emit(p.tokenBuilder.CharacterToken(r))
		if p.tokenBuilder.TempBuffer() == "script" {
			return false, scriptDataEscapedState
		}
		return false, scriptDataDoubleEscapedState
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		r += 0x20
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeEndState
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z':
		p.emit(p.tokenBuilder.CharacterToken(r))
		p.tokenBuilder.WriteTempBuffer(r)
		return false, scriptDataDoubleEscapeEndState
	default:
		return true, scriptDataDoubleEscapedState
	}
}

func (p *HTMLTokenizer) beforeAttributeNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, afterAttributeNameState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeAttributeNameState
	case '/', '>':
		return true, afterAttributeNameState
	case '=':
		// set that attribute's name to the current input character, and its value to the empty string.
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState
	default:
		return true, attributeNameState
	}
}

func (p *HTMLTokenizer) attributeNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.CommitAttribute()
		return true, afterAttributeNameState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020', '/', '>':
		p.tokenBuilder.CommitAttribute()
		return true, afterAttributeNameState
	case '=':
		if p.tokenBuilder.RemoveDuplicateAttributeName() {
			return false, beforeAttributeValueState
		}
		return false, beforeAttributeValueState
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.WriteAttributeName(r + 0x20)
		return false, attributeNameState
	case '\u0000':
		p.tokenBuilder.WriteAttributeName('\uFFFD')
		return false, attributeNameState
	case '"', '\'', '<':
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState
	default:
		p.tokenBuilder.WriteAttributeName(r)
		return false, attributeNameState
	}
}

func (p *HTMLTokenizer) afterAttributeNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterAttributeNameState
	case '/':
		return false, selfClosingStartTagState
	case '=':
		return false, beforeAttributeValueState
	case '>':
		return false, p.emitCurrentTag()
	default:
		return true, attributeNameState
	}
}

func (p *HTMLTokenizer) beforeAttributeValueStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, attributeValueUnquotedState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeAttributeValueState
	case '"':
		return false, attributeValueDoubleQuotedState
	case '\'':
		return false, attributeValueSingleQuotedState
	case '>':
		p.tokenBuilder.CommitAttribute()
		return false, p.emitCurrentTag()
	default:
		return true, attributeValueUnquotedState
	}
}

func (p *HTMLTokenizer) attributeValueDoubleQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '"':
		p.tokenBuilder.CommitAttribute()
		return false, afterAttributeValueQuotedState
	case '&':
		p.returnState = attributeValueDoubleQuotedState
		return false, characterReferenceState
	case '\u0000':
		p.tokenBuilder.WriteAttributeValue('\uFFFD')
		return false, attributeValueDoubleQuotedState
	default:
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueDoubleQuotedState
	}
}

func (p *HTMLTokenizer) attributeValueSingleQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\'':

		p.tokenBuilder.CommitAttribute()
		return false, afterAttributeValueQuotedState
	case '&':
		p.returnState = attributeValueSingleQuotedState
		return false, characterReferenceState
	case '\u0000':
		p.tokenBuilder.WriteAttributeValue('\uFFFD')
		return false, attributeValueSingleQuotedState
	default:
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueSingleQuotedState
	}
}

func (p *HTMLTokenizer) attributeValueUnquotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		p.tokenBuilder.CommitAttribute()
		return false, beforeAttributeNameState
	case '&':
		p.returnState = attributeValueUnquotedState
		return false, characterReferenceState
	case '>':
		p.tokenBuilder.CommitAttribute()
		return false, p.emitCurrentTag()
	case '\u0000':
		p.tokenBuilder.WriteAttributeValue('\uFFFD')
		return false, attributeValueUnquotedState
	case '"', '\'', '<', '=', '`':
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueUnquotedState
	default:
		p.tokenBuilder.WriteAttributeValue(r)
		return false, attributeValueUnquotedState
	}
}

func (p *HTMLTokenizer) afterAttributeValueQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeAttributeNameState
	case '/':
		return false, selfClosingStartTagState
	case '>':
		return false, p.emitCurrentTag()
	default:
		return true, beforeAttributeNameState
	}
}

func (p *HTMLTokenizer) selfClosingStartTagStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '>':
		p.tokenBuilder.EnableSelfClosing()
		return false, p.emitCurrentTag()
	default:
		return true, beforeAttributeNameState
	}
}

func (p *HTMLTokenizer) bogusCommentStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CommentToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState
	case '\u0000':
		p.tokenBuilder.WriteData('\uFFFD')
		return false, bogusCommentState
	default:
		p.tokenBuilder.WriteData(r)
		return false, bogusCommentState
	}
}

// used below to look for peeking at what state to jump to next
var doctype = []byte("octype")
var cdata = []byte("CDATA[")
var peekDist = 6

func (p *HTMLTokenizer) defaultMarkupDeclarationOpenStateParser() (bool, tokenizerState) {
	p.tokenBuilder.Reset()
	return true, bogusCommentState
}

func (p *HTMLTokenizer) markupDeclarationOpenStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.Reset()
		return true, bogusCommentState
	}
	var (
		peeked []byte
		err    error
	)

	switch r {
	case '-':
		peeked, err = p.inputStream.Peek(1)
		if err != nil {
			if len(peeked) < 1 {
				return p.defaultMarkupDeclarationOpenStateParser()
			}
		}
		if len(peeked) == 1 && peeked[0] == '-' {
			p.inputStream.Discard(1)
			p.tokenBuilder.Reset()
			return false, commentStartState
		}

		return p.defaultMarkupDeclarationOpenStateParser()
	case 'D', 'd':
		peeked, err = p.inputStream.Peek(peekDist)
		if err != nil {
			if len(peeked) < peekDist {
				return p.defaultMarkupDeclarationOpenStateParser()
			}
		}
		if bytes.EqualFold(peeked, doctype) {
			p.inputStream.Discard(peekDist)
			return false, doctypeState
		}
	case '[':
		peeked, err = p.inputStream.Peek(peekDist)
		if err != nil {
			if len(peeked) < peekDist {
				return false, bogusCommentState
			}
		}
		if bytes.Equal(cdata, peeked) {
			p.inputStream.Discard(peekDist)
			if p.adjustedCurrentNode != nil && p.adjustedCurrentNode.Element.NamespaceURI != spec.Htmlns {
				return false, cdataSectionState
			}
			p.tokenBuilder.Reset()
			p.tokenBuilder.WriteData('[')
			p.tokenBuilder.WriteData('C')
			p.tokenBuilder.WriteData('D')
			p.tokenBuilder.WriteData('A')
			p.tokenBuilder.WriteData('T')
			p.tokenBuilder.WriteData('A')
			p.tokenBuilder.WriteData('[')
			return false, bogusCommentState
		}
	default:
		return p.defaultMarkupDeclarationOpenStateParser()
	}

	return p.defaultMarkupDeclarationOpenStateParser()
}

func (p *HTMLTokenizer) commentStartStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, commentState
	}
	switch r {
	case '-':
		return false, commentStartDashState
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState
	default:
		return true, commentState
	}
}
func (p *HTMLTokenizer) commentStartDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CommentToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		return false, commentEndState
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState
	default:
		p.tokenBuilder.WriteData('-')
		return true, commentState
	}
}
func (p *HTMLTokenizer) commentStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CommentToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '<':
		p.tokenBuilder.WriteData(r)
		return false, commentLessThanSignState
	case '-':
		return false, commentEndDashState
	case '\u0000':
		p.tokenBuilder.WriteData('\uFFFD')
		return false, commentState
	default:
		p.tokenBuilder.WriteData(r)
		return false, commentState
	}
}
func (p *HTMLTokenizer) commentLessThanSignStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, commentState
	}
	switch r {
	case '!':
		p.tokenBuilder.WriteData(r)
		return false, commentLessThanSignBangState
	case '<':
		p.tokenBuilder.WriteData(r)
		return false, commentLessThanSignState
	default:
		return true, commentState
	}
}
func (p *HTMLTokenizer) commentLessThanSignBangStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, commentState
	}
	switch r {
	case '-':
		return false, commentLessThanSignBangDashState
	default:
		return true, commentState
	}
}
func (p *HTMLTokenizer) commentLessThanSignBangDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, commentEndDashState
	}
	switch r {
	case '-':
		return false, commentLessThanSignBangDashDashState
	default:
		return true, commentEndDashState
	}
}
func (p *HTMLTokenizer) commentLessThanSignBangDashDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, commentEndState
	}
	switch r {
	case '>':
		return true, commentEndState
	default:
		return true, commentEndState
	}
}
func (p *HTMLTokenizer) commentEndDashStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CommentToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		return false, commentEndState
	default:
		p.tokenBuilder.WriteData('-')
		return true, commentState
	}
}
func (p *HTMLTokenizer) commentEndStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {

		p.emit(p.tokenBuilder.CommentToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState
	case '!':
		return false, commentEndBangState
	case '-':
		p.tokenBuilder.WriteData('-')
		return false, commentEndState
	default:
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('-')
		return true, commentState
	}
}
func (p *HTMLTokenizer) commentEndBangStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CommentToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '-':
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('!')
		return false, commentEndDashState
	case '>':
		p.emit(p.tokenBuilder.CommentToken())
		return false, dataState
	default:
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('-')
		p.tokenBuilder.WriteData('!')
		return true, commentState
	}
}
func (p *HTMLTokenizer) doctypeStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.Reset()
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeNameState
	case '>':
		return true, beforeDoctypeNameState
	default:
		return true, beforeDoctypeNameState
	}
}
func (p *HTMLTokenizer) beforeDoctypeNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.Reset()
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeNameState
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		p.tokenBuilder.Reset()
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState
	case '\u0000':
		p.tokenBuilder.Reset()
		p.tokenBuilder.WriteName('\uFFFD')
		return false, doctypeNameState
	case '>':
		p.tokenBuilder.Reset()
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.Reset()
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState
	}
}
func (p *HTMLTokenizer) doctypeNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterDoctypeNameState
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		r += 0x20
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState
	case '\u0000':
		p.tokenBuilder.WriteName('\uFFFD')
		return false, doctypeNameState
	default:
		p.tokenBuilder.WriteName(r)
		return false, doctypeNameState
	}
}
func (p *HTMLTokenizer) afterDoctypeNameStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterDoctypeNameState
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		b, err := p.inputStream.Peek(5)
		if err != nil {
			p.tokenBuilder.EnableForceQuirks()
			return true, bogusDoctypeState
		}
		bs := bytes.Join([][]byte{{byte(r)}, b}, []byte{})
		if bytes.EqualFold(bs, []byte("PUBLIC")) {
			p.inputStream.Discard(5)
			return false, afterDoctypePublicKeywordState
		} else if bytes.EqualFold(bs, []byte("SYSTEM")) {
			p.inputStream.Discard(5)
			return false, afterDoctypeSystemKeywordState
		}
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) afterDoctypePublicKeywordStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypePublicIdentifierState
	case '"':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierDoubleQuotedState
	case '\'':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierSingleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) beforeDoctypePublicIdentifierStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypePublicIdentifierState
	case '"':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierDoubleQuotedState
	case '\'':
		p.tokenBuilder.WritePublicIdentifierEmpty()
		return false, doctypePublicIdentifierSingleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) doctypePublicIdentifierDoubleQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '"':
		return false, afterDoctypePublicIdentifierState
	case '\u0000':
		p.tokenBuilder.WritePublicIdentifier('\uFFFD')
		return false, doctypePublicIdentifierDoubleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.WritePublicIdentifier(r)
		return false, doctypePublicIdentifierDoubleQuotedState
	}
}
func (p *HTMLTokenizer) doctypePublicIdentifierSingleQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\'':
		return false, afterDoctypePublicIdentifierState
	case '\u0000':
		p.tokenBuilder.WritePublicIdentifier('\uFFFD')
		return false, doctypePublicIdentifierSingleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.WritePublicIdentifier(r)
		return false, doctypePublicIdentifierSingleQuotedState
	}
}
func (p *HTMLTokenizer) afterDoctypePublicIdentifierStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, betweenDoctypePublicAndSystemIdentifiersState
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) betweenDoctypePublicAndSystemIdentifiersStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, betweenDoctypePublicAndSystemIdentifiersState
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}

func (p *HTMLTokenizer) afterDoctypeSystemKeywordStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeSystemIdentifierState
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}

func (p *HTMLTokenizer) beforeDoctypeSystemIdentifierStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, beforeDoctypeSystemIdentifierState
	case '"':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierDoubleQuotedState
	case '\'':
		p.tokenBuilder.WriteSystemIdentifierEmpty()
		return false, doctypeSystemIdentifierSingleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.EnableForceQuirks()
		return true, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) doctypeSystemIdentifierDoubleQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '"':
		return false, afterDoctypeSystemIdentifierState
	case '\u0000':
		p.tokenBuilder.WriteSystemIdentifier('\uFFFD')
		return false, doctypeSystemIdentifierDoubleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.WriteSystemIdentifier(r)
		return false, doctypeSystemIdentifierDoubleQuotedState
	}
}
func (p *HTMLTokenizer) doctypeSystemIdentifierSingleQuotedStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\'':
		return false, afterDoctypeSystemIdentifierState
	case '\u0000':
		p.tokenBuilder.WriteSystemIdentifier('\uFFFD')
		return false, doctypeSystemIdentifierSingleQuotedState
	case '>':
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		p.tokenBuilder.WriteSystemIdentifier(r)
		return false, doctypeSystemIdentifierSingleQuotedState
	}
}
func (p *HTMLTokenizer) afterDoctypeSystemIdentifierStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.tokenBuilder.EnableForceQuirks()
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '\u0009', '\u000A', '\u000C', '\u0020':
		return false, afterDoctypeSystemIdentifierState
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	default:
		return true, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) bogusDoctypeStateParser(r rune, eof bool) (bool, tokenizerState) {

	if eof {
		p.emit(p.tokenBuilder.DocTypeToken(), p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case '>':
		p.emit(p.tokenBuilder.DocTypeToken())
		return false, dataState
	case '\u0000':
		return false, bogusDoctypeState
	default:
		return false, bogusDoctypeState
	}
}
func (p *HTMLTokenizer) cdataSectionStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.EndOfFileToken())
		return false, dataState
	}
	switch r {
	case ']':
		return false, cdataSectionBracketState
	default:
		p.emit(p.tokenBuilder.CharacterToken(r))
		return false, cdataSectionState
	}
}
func (p *HTMLTokenizer) cdataSectionBracketStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState
	}
	switch r {
	case ']':
		return false, cdataSectionEndState
	default:
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState
	}
}
func (p *HTMLTokenizer) cdataSectionEndStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.emit(p.tokenBuilder.CharacterToken(']'), p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState
	}
	switch r {
	case ']':
		p.emit(p.tokenBuilder.CharacterToken(']'))
		return false, cdataSectionEndState
	case '>':
		return false, dataState
	default:
		p.emit(p.tokenBuilder.CharacterToken(']'), p.tokenBuilder.CharacterToken(']'))
		return true, cdataSectionState
	}
}

func (p *HTMLTokenizer) characterReferenceStateParser(r rune, eof bool) (bool, tokenizerState) {
	p.tokenBuilder.ResetTempBuffer()
	p.tokenBuilder.WriteTempBuffer('&')

	if eof {
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true, namedCharacterReferenceState
	case '#':
		p.tokenBuilder.WriteTempBuffer(r)
		return false, numericCharacterReferenceState
	default:
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState
	}
}

func (p *HTMLTokenizer) anyFilteredTable(filteredTable map[string][]rune, match string) (bool, string) {
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

func (p *HTMLTokenizer) namedCharacterReferenceStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.flushCodePointsAsCharacterReference()
		return false, ambiguousAmpersandState
	}

	filteredTable := map[string][]rune{}
	for k, v := range charRefTable {
		filteredTable[k] = v
	}

	consumed := []byte{byte(r)}
	hasMatches, smallest := p.anyFilteredTable(filteredTable, string(consumed))

	var (
		b, sep, next []byte
		err          error
		newSmallest  string
		firstDiscard bool = true
	)
	match := bytes.Join([][]byte{consumed, b}, sep)

	// consume one character until there are no more matches in the filtered table
	for i := 1; hasMatches; i++ {
		b, err = p.inputStream.Peek(i)
		// if we can't peek any more bytes, leave the loop and use the currently
		// calculated smallest match as the match.
		if err != nil {
			break
		}

		// concatenate the read bytes to the consumed
		match = bytes.Join([][]byte{consumed, b}, sep)
		hasMatches, newSmallest = p.anyFilteredTable(filteredTable, string(match))

		// if we found a larger, better match in the table, update the reader and
		// the loop counter
		if len(newSmallest) > len(smallest) {
			i = 0
			diff := len(newSmallest) - len(smallest)

			// because we already dicarded the first rune in the main loop
			// we don't need discard it again here.
			if firstDiscard {
				_, err = p.inputStream.Discard(diff - 1)
				firstDiscard = false
			} else {
				_, err = p.inputStream.Discard(diff)
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
				p.inputStream.Discard(1)
			}
			p.tokenBuilder.WriteTempBuffer(rune(mb))
		}
		p.flushCodePointsAsCharacterReference()
		return false, ambiguousAmpersandState
	}

	endsInSemiColon := bytes.HasSuffix(consumed, []byte{';'})
	if wasConsumedByAttribute(p.returnState) && !endsInSemiColon {
		next, err = p.inputStream.Peek(1)
		if err == nil {
			switch next[0] {
			case '=', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				for _, r := range next {
					p.tokenBuilder.WriteTempBuffer(rune(r))
					p.inputStream.Discard(1)
				}
				p.flushCodePointsAsCharacterReference()
				return false, p.returnState
			}
		}
	}
	p.tokenBuilder.ResetTempBuffer()
	for _, r := range charRefTable[string(consumed)] {
		p.tokenBuilder.WriteTempBuffer(r)
	}
	p.flushCodePointsAsCharacterReference()

	return false, p.returnState
}

func (p *HTMLTokenizer) ambiguousAmpersandStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, p.returnState
	}
	switch r {
	case 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if wasConsumedByAttribute(p.returnState) {
			p.tokenBuilder.WriteAttributeValue(r)
		} else {
			p.emit(p.tokenBuilder.CharacterToken(r))
		}
		return false, ambiguousAmpersandState
	case ';':
		return true, p.returnState
	default:
		return true, p.returnState
	}
}
func (p *HTMLTokenizer) numericCharacterReferenceStateParser(r rune, eof bool) (bool, tokenizerState) {
	p.tokenBuilder.SetCharRef(0)
	if eof {
		return true, decimalCharacterReferenceStartState
	}
	switch r {
	case 'x', 'X':
		p.tokenBuilder.WriteTempBuffer(r)
		return false, hexadecimalCharacterReferenceStartState
	default:
		return true, decimalCharacterReferenceStartState
	}
}

func (p *HTMLTokenizer) hexadecimalCharacterReferenceStartStateParser(r rune, eof bool) (bool, tokenizerState) {

	if eof {
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'a', 'b', 'c', 'd', 'e', 'f':
		return true, hexadecimalCharacterReferenceState
	default:
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState
	}
}

func (p *HTMLTokenizer) decimalCharacterReferenceStartStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true, decimalCharacterReferenceState
	default:
		p.flushCodePointsAsCharacterReference()
		return true, p.returnState
	}
}

func (p *HTMLTokenizer) hexadecimalCharacterReferenceStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, numericCharacterReferenceEndState
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.tokenBuilder.MultByCharRef(16)
		p.tokenBuilder.AddToCharRef(int(r - 0x30))
		return false, hexadecimalCharacterReferenceState
	case 'A', 'B', 'C', 'D', 'E', 'F':
		p.tokenBuilder.MultByCharRef(16)
		p.tokenBuilder.AddToCharRef(int(r - 0x37))
		return false, hexadecimalCharacterReferenceState
	case 'a', 'b', 'c', 'd', 'e', 'f':
		p.tokenBuilder.MultByCharRef(16)
		p.tokenBuilder.AddToCharRef(int(r - 0x57))
		return false, hexadecimalCharacterReferenceState
	case ';':
		return false, numericCharacterReferenceEndState
	default:
		return true, numericCharacterReferenceEndState
	}
}

func (p *HTMLTokenizer) decimalCharacterReferenceStateParser(r rune, eof bool) (bool, tokenizerState) {
	if eof {
		return true, numericCharacterReferenceEndState
	}
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		p.tokenBuilder.MultByCharRef(10)
		p.tokenBuilder.AddToCharRef(int(r - 0x30))
		return false, decimalCharacterReferenceState
	case ';':
		return false, numericCharacterReferenceEndState
	default:
		return true, numericCharacterReferenceEndState
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

func (p *HTMLTokenizer) numericCharacterReferenceEndStateParser(r rune, eof bool) (bool, tokenizerState) {
	// this is the only state that isn't suppose to consume something off the bat.
	p.inputStream.UnreadRune()
	if p.tokenBuilder.Cmp(0) == 0 {
		p.tokenBuilder.SetCharRef(0xFFFD)
	} else if p.tokenBuilder.Cmp(0x10FFFF) == 1 {
		p.tokenBuilder.SetCharRef(0xFFFD)
	} else if isSurrogate(p.tokenBuilder.GetCharRef()) {
		p.tokenBuilder.SetCharRef(0xFFFD)
	} else if isNonCharacter(p.tokenBuilder.GetCharRef()) {
	} else if p.tokenBuilder.Cmp(0x0D) == 0 ||
		(isControl(p.tokenBuilder.GetCharRef())) && !isASCIIWhitespace(p.tokenBuilder.GetCharRef()) {
		tableNum, ok := numericCharacterReferenceEndStateTable[p.tokenBuilder.GetCharRef()]
		if ok {
			p.tokenBuilder.SetCharRef(int(tableNum))
		}
	}

	p.tokenBuilder.ResetTempBuffer()
	p.tokenBuilder.WriteTempBuffer(rune(p.tokenBuilder.GetCharRef()))
	p.flushCodePointsAsCharacterReference()
	return false, p.returnState
}

func (p *HTMLTokenizer) emitCurrentTag() tokenizerState {
	switch p.tokenBuilder.curTagType {
	case startTag:
		p.emit(p.tokenBuilder.StartTagToken())
	case endTag:
		p.emit(p.tokenBuilder.EndTagToken())
	}

	return dataState
}

// a stateHandler is a func that takes in a rune and a bool representing the endoffile
// and returns the next state to transition to.
type parserStateHandler func(in rune, eof bool) (bool, tokenizerState)

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

func (p *HTMLTokenizer) normalizeNewlines(r rune) rune {
	if r == '\u000D' {
		b, err := p.inputStream.Peek(1)
		if err != nil {
			return '\u000A'
		}
		if len(b) > 0 && b[0] == '\u000A' {
			p.inputStream.Discard(1)
		}

		return '\u000A'
	}

	return r
}

func (p *HTMLTokenizer) takeLastEmittedToken() *Token {
	if len(p.emittedTokens) > 0 {
		ret := p.emittedTokens[0]
		p.emittedTokens = p.emittedTokens[1:]
		if ret.TokenType == endOfFileToken {
			p.done = true
		}
		return &ret
	}
	return nil
}

func (p *HTMLTokenizer) Next() bool {
	return !p.done
}

func (p *HTMLTokenizer) Token(progress *Progress) (*Token, error) {
	// the tree constructor needs to be able to change the state of the tokenizer.
	// TODO: this only is in certain cases. if we can hide this behind a feature flag
	// we can make the normal case a bit more abstracted between tokenizer and tree
	// constructor.
	// if TokenizerState is set, the tree constructor set it.
	p.adjustedCurrentNode = progress.AdjustedCurrentNode
	if progress.TokenizerState != nil {
		p.currentState = *progress.TokenizerState
	}

	// some states emit more than 1 token at a time and sometimes no tokens.
	// loop until at least 1 token is emitted and then take them.
	for {
		token := p.takeLastEmittedToken()
		if token != nil {
			return token, nil
		}

		r, _, err := p.inputStream.ReadRune()
		if err != nil && err != io.EOF {
			return nil, err
		}

		p.processRune(p.normalizeNewlines(r), err == io.EOF)
	}
}

func (p *HTMLTokenizer) processRune(r rune, eof bool) {
	reconsume := true
	for reconsume {
		reconsume, p.currentState = p.stateToParser(p.currentState)(r, eof)
		fmt.Printf("[TOKEN]rune: %s , mode: %s\n", string(r), p.currentState)
	}
}
