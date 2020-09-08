package html

import "browser/parser/dom"

type HTMLTemplateElement struct {
	content dom.DocumentFragment
	*HTMLElement
}
