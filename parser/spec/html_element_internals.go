package spec

import (
	"browser/parser/webidl"
)

type ElementInternals struct {
	form              *HTMLForm
	willValidate      bool
	validity          ValidityState
	validationMessage webidl.DOMString
	labels            NodeList
}

func (ei *ElementInternals) setFormValue(value, state webidl.USVString) {
}
func (ei *ElementInternals) setValidity(flags ValidityStateFlags, message webidl.DOMString, anchor *HTMLElement) {
}
func (ei *ElementInternals) checkValidity() bool  { return false }
func (ei *ElementInternals) reportValidity() bool { return false }
