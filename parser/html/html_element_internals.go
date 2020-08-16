package html

import (
	"browser/parser/dom"
	"browser/parser/webidl"
)

type ElementInternals struct {
	form              *HTMLFormElement
	willValidate      bool
	validity          ValidityState
	validationMessage webidl.DOMString
	labels            dom.NodeList
}

func (ei *ElementInternals) setFormValue(value, state webidl.USVString) {
}
func (ei *ElementInternals) setValidity(flags ValidityStateFlags, message webidl.DOMString, anchor *HTMLELement) {
}
func (ei *ElementInternals) checkValidity() bool  { return false }
func (ei *ElementInternals) reportValidity() bool { return false }
