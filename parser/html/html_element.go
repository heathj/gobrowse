package html

import (
	"browser/parser/dom"
	"browser/parser/webidl"
)

type HTMLELement struct {
	title, lang, dir, accessKey, accessKeyLabel, autocapitalize, innerText webidl.DOMString
	translate, hidden, draggable, spellcheck                               bool

	*dom.Element
	*GlobalEventHandlers
	*DocumentAndElementEventHandlers
	*ElementContentEditable
	*HTMLOrSVGElement
}

func (e *HTMLELement) attachInternals() ElementInternals { return nil }
func (e *HTMLELement) click()                            {}
