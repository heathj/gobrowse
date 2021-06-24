package spec

// https:domspec.whatwg.org/#dictdef-elementcreationoptions
type ElementCreationOptions map[elementCreationKeys]string
type elementCreationKeys uint

const (
	is elementCreationKeys = iota
)
