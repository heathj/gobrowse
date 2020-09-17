package spec

// https:domspec.whatwg.org/#nodeiterator
type NodeIterator struct {
	root                       *Node
	referenceNode              *Node
	pointerBeforeReferenceNode bool
	whatToShow                 uint
	filter                     NodeFilter
}

func (n *NodeIterator) nextNode() *Node     { return nil }
func (n *NodeIterator) previousNode() *Node { return nil }
func (n *NodeIterator) detach()             {}
