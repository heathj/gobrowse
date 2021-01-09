package parser

import (
	"fmt"
	"math/big"
	"strings"
)

//go:generate stringer -type=tokenType
type tokenType uint

const (
	characterToken tokenType = iota
	startTagToken
	endTagToken
	endOfFileToken
	commentToken
	docTypeToken
)

const missing string = "MISSING"

type tagType uint

const (
	startTag tagType = iota
	endTag
)

// Token is a concrete token that is ready to be emitted.
type Token struct {
	TokenType        tokenType
	Attributes       map[string]string
	TagName          string
	PublicIdentifier string
	SystemIdentifier string
	ForceQuirks      bool
	SelfClosing      bool
	Data             string
	Special          bool
}

func (t *Token) String() string {
	switch t.TokenType {
	case characterToken, commentToken:
		return fmt.Sprintf(`Token: %s:
	Data: %q
`, t.TokenType, t.Data)
	case startTagToken, endTagToken:
		return fmt.Sprintf(`Token: %s:
	TagName: %q
	Attributes: %+v
	SelfClosing: %t
`, t.TokenType, t.TagName, t.Attributes, t.SelfClosing)
	case endOfFileToken:
		return `Token: EofToken
`
	case docTypeToken:
		return fmt.Sprintf(`Token: DOCTYPE token
	TagName: %q
	ForQuirks: %t
	PublicID: %q
	SystemID: %q
`, t.TagName, t.ForceQuirks, t.PublicIdentifier, t.SystemIdentifier)
	}

	return ""
}

// Equal compares if two tokens are equal to each other.
func (a *Token) Equal(b *Token) bool {
	if a.TokenType != b.TokenType {
		return false
	}

	switch a.TokenType {
	case characterToken, commentToken:
		if a.Data != b.Data {
			return false
		}
	case startTagToken, endTagToken:
		if a.TagName != b.TagName {
			return false
		}

		if a.SelfClosing != b.SelfClosing {
			return false
		}

		if len(a.Attributes) != len(b.Attributes) {
			return false
		}
		for k := range a.Attributes {
			if _, ok := b.Attributes[k]; !ok {
				return false
			}

			if a.Attributes[k] != b.Attributes[k] {
				return false
			}
		}
	case docTypeToken:
		if a.TagName != b.TagName {
			return false
		}
		if a.ForceQuirks != b.ForceQuirks {
			return false
		}
		if a.PublicIdentifier != b.PublicIdentifier {
			return false
		}
		if a.SystemIdentifier != b.SystemIdentifier {
			return false
		}
	}
	return true
}

// TokenBuilder builds various tokens up during the tokenization
// phase.
type TokenBuilder struct {
	attributes             map[string]string
	attributeKey           strings.Builder
	attributeValue         strings.Builder
	name                   strings.Builder
	data                   strings.Builder
	tempBuffer             strings.Builder
	publicID               strings.Builder
	systemID               strings.Builder
	selfClosing            bool
	forceQuirks            bool
	removeNextAttr         bool
	curTagType             tagType
	characterReferenceCode *big.Int
}

func newTokenBuilder() *TokenBuilder {
	return &TokenBuilder{
		attributes:             make(map[string]string),
		characterReferenceCode: big.NewInt(0),
	}
}

// NewToken clears all the builders and attributes. We don't include
// the temp buffer here because I am not sure where I need to clear that one yet.
func (t *TokenBuilder) NewToken() {
	t.attributes = make(map[string]string)
	t.attributeKey.Reset()
	t.attributeValue.Reset()
	//default state for public and system id is "MISSING"
	t.publicID.Reset()
	t.systemID.Reset()
	t.publicID.WriteString(missing)
	t.systemID.WriteString(missing)
	t.data.Reset()
	t.name.Reset()
	t.selfClosing = false
	t.forceQuirks = false
	t.removeNextAttr = false
	t.characterReferenceCode = big.NewInt(0)
}

// EnableSelfClosing changes to the self-closing flag to "set".
func (t *TokenBuilder) EnableSelfClosing() {
	t.selfClosing = true
}

// EnableForceQuirks changes to the force-quirks flag to "set".
func (t *TokenBuilder) EnableForceQuirks() {
	t.forceQuirks = true
}

// WritePublicIdentifier appends a rune to the public identifier buffer.
func (t *TokenBuilder) WritePublicIdentifier(r rune) {
	if t.publicID.String() == missing {
		t.publicID.Reset()
	}
	_, err := t.publicID.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

//WritePublicIdentifierEmpty writes the empty string to the public ID.
func (t *TokenBuilder) WritePublicIdentifierEmpty() {
	t.publicID.Reset()
}

//WriteSystemIdentifierEmpty writes the empty string to the system ID.
func (t *TokenBuilder) WriteSystemIdentifierEmpty() {
	t.systemID.Reset()
}

// WriteSystemIdentifier appends a rune to the public identifier buffer.
func (t *TokenBuilder) WriteSystemIdentifier(r rune) {
	if t.systemID.String() == missing {
		t.systemID.Reset()
	}
	_, err := t.systemID.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

// WriteAttributeName appends a character to the current
// attribute's name.
func (t *TokenBuilder) WriteAttributeName(r rune) {
	_, err := t.attributeKey.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

//WriteData appends a character to the current data section.
func (t *TokenBuilder) WriteData(r rune) {
	_, err := t.data.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

// WriteAttributeValue appends a character to the current
// attribute's value.
func (t *TokenBuilder) WriteAttributeValue(r rune) {
	_, err := t.attributeValue.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

// RemoveDuplicateAttributeName checks if the current name is already
// in the list of commited attributes. If so, it removes the attribute.
func (t *TokenBuilder) RemoveDuplicateAttributeName() bool {
	_, ok := t.attributes[t.attributeKey.String()]
	if ok {
		t.removeNextAttr = true
	}
	return ok
}

// WriteName appends a character to the current name value.
func (t *TokenBuilder) WriteName(r rune) {
	_, err := t.name.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

// CommitAttribute ends the creation of a key/value
// pair by copying the name and value fields into the
// attribute field and clearing the name and value fields.
func (t *TokenBuilder) CommitAttribute() {
	// only commit the attribute if it isn't a duplicate
	if !t.removeNextAttr {
		k := t.attributeKey.String()
		v := t.attributeValue.String()

		if k != "" {
			t.attributes[k] = v
		}
	}
	t.attributeKey.Reset()
	t.attributeValue.Reset()
	t.removeNextAttr = false
}

// WriteTempBuffer appends a character to the temporary buffer of the current
// state.
func (t *TokenBuilder) WriteTempBuffer(r rune) {
	_, err := t.tempBuffer.WriteRune(r)
	if err != nil {
		fmt.Print(err)
	}
}

// ResetTempBuffer clears the temporary buffer to be used by some other state.
func (t *TokenBuilder) ResetTempBuffer() {
	t.tempBuffer.Reset()
}

// TempBuffer just returns the string version of the current buffer conents.
func (t *TokenBuilder) TempBuffer() string {
	return t.tempBuffer.String()
}

// SetCharRef sets the internal character reference count to zero.
func (t *TokenBuilder) SetCharRef(i int) {
	t.characterReferenceCode = big.NewInt(int64(i))
}

// GetCharRef sets the internal character reference count to zero.
func (t *TokenBuilder) GetCharRef() int {
	return int(t.characterReferenceCode.Int64())
}

func (t *TokenBuilder) Cmp(i int) int {
	return t.characterReferenceCode.Cmp(big.NewInt(int64(i)))
}

// AddToCharRef adds a number to the current char ref count.
func (t *TokenBuilder) AddToCharRef(i int) {
	t.characterReferenceCode.Add(t.characterReferenceCode, big.NewInt(int64(i)))
}

// MultByCharRef multiplys a number to the current char ref count.
func (t *TokenBuilder) MultByCharRef(i int) {
	t.characterReferenceCode.Mul(t.characterReferenceCode, big.NewInt(int64(i)))

}

// StartTagToken creates a start tag token from the builder
// contents.
func (t *TokenBuilder) StartTagToken() Token {
	return Token{
		TokenType:   startTagToken,
		TagName:     t.name.String(),
		Attributes:  t.attributes,
		SelfClosing: t.selfClosing,
	}
}

// EndTagToken creates an end tag token from the builder
// contents.
func (t *TokenBuilder) EndTagToken() Token {
	return Token{
		TokenType:   endTagToken,
		TagName:     t.name.String(),
		Attributes:  t.attributes,
		SelfClosing: t.selfClosing,
	}
}

// CharacterToken creates a character token from the builder
// contents.
func (t *TokenBuilder) CharacterToken(r rune) Token {
	return Token{
		TokenType: characterToken,
		Data:      string(r),
	}
}

// EndOfFileToken create an end of file token.
func (t *TokenBuilder) EndOfFileToken() Token {
	return Token{
		TokenType: endOfFileToken,
	}
}

// CommentToken creates a comment token from the builder contents.
func (t *TokenBuilder) CommentToken() Token {
	return Token{
		TokenType: commentToken,
		Data:      t.data.String(),
	}
}

// DocTypeToken creates a doc type token from the builder contents.
func (t *TokenBuilder) DocTypeToken() Token {
	return Token{
		TokenType:        docTypeToken,
		TagName:          t.name.String(),
		ForceQuirks:      t.forceQuirks,
		PublicIdentifier: t.publicID.String(),
		SystemIdentifier: t.systemID.String(),
	}
}
