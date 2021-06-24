package spec

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
	Location                                                       *HTMLLocation
	Domain, Referrer, Cookie, LastModified, Title, Dir, DesignMode string
	ReadyState                                                     DocumentReadyState
	Body                                                           *HTMLElement
	Head                                                           *HTMLHead
	Images, Embeds, Plugins, Links, Forms, Scripts                 HTMLCollection
	CurrentScript                                                  HTMLOrSVGScript
	DefaultView                                                    *WindowProxy
	Onreadystatechange                                             EventHandler

	*Node
}

func (d *HTMLDocument) GetElementsByName(elementName string) NodeList { return nil }
func (d *HTMLDocument) Open(u1, u2 string) *HTMLDocument              { return nil }
func (d *HTMLDocument) OpenW(url string, name, features string) *WindowProxy {
	return nil
}
func (d *HTMLDocument) Close()                 {}
func (d *HTMLDocument) Write(text ...string)   {}
func (d *HTMLDocument) Writeln(text ...string) {}
func (d *HTMLDocument) HasFocus() bool         { return false }
func (d *HTMLDocument) ExecCommand(commandID string, showUI bool, value string) bool {
	return false
}
func (d *HTMLDocument) QueryCommandEnabled(commandID string) bool {
	return false
}
func (d *HTMLDocument) QueryCommandIndeterm(commandID string) bool {
	return false
}
func (d *HTMLDocument) QueryCommandState(commandID string) bool {
	return false
}
func (d *HTMLDocument) QueryCommandSupported(commandID string) bool {
	return false
}
func (d *HTMLDocument) QueryCommandValue(commandID string) string {
	return ""
}
