package html

import (
	"browser/parser/dom"
	"browser/parser/webidl"
)

type HTMLElement struct {
	Title, Lang, Dir, AccessKey, AccessKeyLabel, Autocapitalize, InnerText webidl.DOMString
	Translate, Hidden, Draggable, Spellcheck                               bool

	*dom.Element
	*GlobalEventHandlers
	*DocumentAndElementEventHandlers
	*ElementContentEditable
	*HTMLOrSVGElement
}

func (e *HTMLElement) attachInternals() *ElementInternals { return nil }
func (e *HTMLElement) click()                             {}
