package spec

import "browser/parser/webidl"

type DOMStringList []webidl.DOMString
type HTMLLocation struct {
	href            webidl.USVString
	origin          webidl.USVString
	protocol        webidl.USVString
	host            webidl.USVString
	hostname        webidl.USVString
	port            webidl.USVString
	pathname        webidl.USVString
	search          webidl.USVString
	hash            webidl.USVString
	ancestorOrigins DOMStringList
}

func (l *HTMLLocation) assign(url webidl.USVString)  {}
func (l *HTMLLocation) replace(url webidl.USVString) {}
func (l *HTMLLocation) reload()                      {}
