package spec

type ElementInternals struct {
	form              *HTMLForm
	willValidate      bool
	validity          ValidityState
	validationMessage string
	labels            NodeList
}

func (ei *ElementInternals) setFormValue(value, state string) {
}
func (ei *ElementInternals) setValidity(flags ValidityStateFlags, message string, anchor *HTMLElement) {
}
func (ei *ElementInternals) checkValidity() bool  { return false }
func (ei *ElementInternals) reportValidity() bool { return false }
