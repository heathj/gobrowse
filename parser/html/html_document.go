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
	Location           *HTMLLocation
	Domain             webidl.USVString
	Referrer           webidl.USVString
	Cookie             webidl.USVString
	LastModified       webidl.DOMString
	ReadyState         DocumentReadyState
	Title              webidl.DOMString
	Dir                webidl.DOMString
	Body               *HTMLElement
	Head               *HTMLHead
	Images             dom.HTMLCollection
	Embeds             dom.HTMLCollection
	Plugins            dom.HTMLCollection
	Links              dom.HTMLCollection
	Forms              dom.HTMLCollection
	Scripts            dom.HTMLCollection
	CurrentScript      HTMLOrSVGScript
	DefaultView        *WindowProxy
	DesignMode         webidl.DOMString
	Onreadystatechange EventHandler

	*dom.Node
}

func (d *HTMLDocument) GetElementsByName(elementName webidl.DOMString) dom.NodeList { return nil }
func (d *HTMLDocument) Open(u1, u2 webidl.DOMString) *HTMLDocument                  { return nil }
func (d *HTMLDocument) OpenW(url webidl.USVString, name, features webidl.DOMString) *WindowProxy {
	return nil
}
func (d *HTMLDocument) Close()                           {}
func (d *HTMLDocument) Write(text ...webidl.DOMString)   {}
func (d *HTMLDocument) Writeln(text ...webidl.DOMString) {}
func (d *HTMLDocument) HasFocus() bool                   { return false }
func (d *HTMLDocument) ExecCommand(commandID webidl.DOMString, showUI bool, value webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) QueryCommandEnabled(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) QueryCommandIndeterm(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) QueryCommandState(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) QueryCommandSupported(commandID webidl.DOMString) bool {
	return false
}
func (d *HTMLDocument) QueryCommandValue(commandID webidl.DOMString) webidl.DOMString {
	return ""
}
