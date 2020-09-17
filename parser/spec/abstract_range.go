package spec

// AbstractRange is https:domspec.whatwg.org/#abstractrange
type AbstractRange struct {
	startContainer *Node
	startOffset    uint
	endContainer   *Node
	endOffSet      uint
	collapsed      bool
}
