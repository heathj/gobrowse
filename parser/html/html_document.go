package html

import (
	"browser/parser/dom"
	"browser/parser/webidl"
)

type DocumentReadyState string

const (
	loading     DocumentReadyState = "loading"
	interactive DocumentReadyState = "interactive"
	complete    DocumentReadyState = "complete"
)

//https://html.spec.whatwg.org/#eventhandler
type EventHandler func(e *dom.Event)

// https://html.spec.whatwg.org/#the-document-object
type HTMLDocument struct {
	location           *HTMLLocation
	domain             webidl.USVString
	referrer           webidl.USVString
	cookie             webidl.USVString
	lastModified       webidl.DOMString
	readyState         DocumentReadyState
	title              webidl.DOMString
	dir                webidl.DOMString
	body               *HTMLElement
	head               *HTMHeadElement
	images             dom.HTMLCollection
	embeds             dom.HTMLCollection
	plugins            dom.HTMLCollection
	links              dom.HTMLCollection
	forms              dom.HTMLCollection
	scripts            dom.HTMLCollection
	currentScript      HTMLOrSVGScriptElement
	defaultView        *WindowProxy
	designMode         webidl.DOMString
	onreadystatechange EventHandler
	*dom.Document
}

func (d *HTMLDocument) getElementsByName(elementName webidl.DOMString) dom.NodeList { return nil }
func (d *HTMLDocument) open(u1, u2 webidl.DOMString) *HTMLDocument                  { return nil }
func (d *HTMLDocument) openW(url webidl.USVString, name, features webidl.DOMString) *WindowProxy {
	return nil
}
func (d *HTMLDocument) close()                           {}
func (d *HTMLDocument) write(text ...webidl.DOMString)   {}
func (d *HTMLDocument) writeln(text ...webidl.DOMString) {}
func (d *HTMLDocument) hasFocus() bool                   { return false }
func (d *HTMLDocument) execCommand(commandID webidl.DOMString, showUI bool, value webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) queryCommandEnabled(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) queryCommandIndeterm(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) queryCommandState(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) queryCommandSupported(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) queryCommandValue(commandID webidl.DOMString) webidl.DOMString {
	return ""
}
