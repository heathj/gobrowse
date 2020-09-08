package dom

// AbstractRange is https://dom.spec.whatwg.org/#abstractrange
type AbstractRange struct {
	startContainer *Node
	startOffset    uint
	endContainer   *Node
	endOffSet      uint
	collapsed      bool
}
