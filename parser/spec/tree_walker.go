package spec

// https:domspec.whatwg.org/#treewalker
type TreeWalker struct {
	root        *Node
	whatToShow  uint
	filter      NodeFilter
	currentNode *Node
}

func (t *TreeWalker) parentNode() *Node      { return nil }
func (t *TreeWalker) firstChild() *Node      { return nil }
func (t *TreeWalker) lastChild() *Node       { return nil }
func (t *TreeWalker) previousSibling() *Node { return nil }
func (t *TreeWalker) nextSibling() *Node     { return nil }
func (t *TreeWalker) previousNode() *Node    { return nil }
func (t *TreeWalker) nextNode() *Node        { return nil }
