package dom

// https://dom.spec.whatwg.org/#dictdef-elementcreationoptions
type ElementCreationOptions map[elementCreationKeys]DOMString
type elementCreationKeys uint

const (
	is elementCreationKeys = iota
)
