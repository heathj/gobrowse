package spec

type DOMStringList []string
type HTMLLocation struct {
	href, origin, protocol, host, hostname, port, pathname, search, hash string
	ancestorOrigins                                                      DOMStringList
}

func (l *HTMLLocation) assign(url string)  {}
func (l *HTMLLocation) replace(url string) {}
func (l *HTMLLocation) reload()            {}
