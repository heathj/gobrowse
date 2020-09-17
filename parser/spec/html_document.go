package spec

import (
	"browser/parser/webidl"
)

type DocumentReadyState string

const (
	loading     DocumentReadyState = "loading"
	interactive DocumentReadyState = "interactive"
	complete    DocumentReadyState = "complete"
)

//https://html.spec.whatwg.org/#eventhandler
type EventHandler func(e *Event)

// https://html.spec.whatwg.org/#the-document-object
type HTMLDocument struct {
	Location                                       *HTMLLocation
	Domain, Referrer, Cookie                       webidl.USVString
	LastModified, Title, Dir, DesignMode           webidl.DOMString
	ReadyState                                     DocumentReadyState
	Body                                           *HTMLElement
	Head                                           *HTMLHead
	Images, Embeds, Plugins, Links, Forms, Scripts HTMLCollection
	CurrentScript                                  HTMLOrSVGScript
	DefaultView                                    *WindowProxy
	Onreadystatechange                             EventHandler

	*Node
}

func (d *HTMLDocument) GetElementsByName(elementName webidl.DOMString) NodeList { return nil }
func (d *HTMLDocument) Open(u1, u2 webidl.DOMString) *HTMLDocument              { return nil }
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
