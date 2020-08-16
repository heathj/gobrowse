package dom

// https://dom.spec.whatwg.org/#htmlcollection
type HTMLCollection []Element

// Element is an individual HTML element that gets added to the DOM.
// https://dom.spec.whatwg.org/#interface-element
type Element struct {
	elemType elementType
	children []*Element

	*NodeFields
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
