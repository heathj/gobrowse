package spec

// https:domspec.whatwg.org/#nodelist
type NodeList []*Node
type NodeRewinder struct {
	nodeList NodeList
	i        int
}

func (n *NodeRewinder) Prev() bool {
	return n.i >= 0
}

func (n *NodeRewinder) Node() *Node {
	if n.i >= 0 && n.i < len(n.nodeList) {
		node := n.nodeList[n.i]
		n.i--
		return node
	}
	return nil
}

func NewNodeRewinder(nl NodeList) *NodeRewinder {
	return &NodeRewinder{
		nodeList: nl,
		i:        len(nl) - 1,
	}
}

func (n *NodeRewinder) WithStart(i int) *NodeRewinder {
	n.i = i
	return n
}

type NodeIterator struct {
	nodeList NodeList
	i        int
}

func (n *NodeIterator) Next() bool {
	return n.i < len(n.nodeList)
}

func (n *NodeIterator) Node() *Node {
	if n.i >= 0 && n.i < len(n.nodeList) {
		node := n.nodeList[n.i]
		n.i++
		return node
	}
	return nil
}

func NewNodeIterator(nl NodeList) *NodeIterator {
	return &NodeIterator{
		nodeList: nl,
		i:        0,
	}
}

func (n *NodeIterator) WithStart(i int) *NodeIterator {
	n.i = i
	return n
}

func (n *NodeIterator) WithStartFrom(sn *Node) *NodeIterator {
	if sn == nil {
		return n
	}
	i := n.nodeList.Contains(sn)

	if i == -1 {
		return n
	}
	n.i = i
	return n
}

func (h *NodeList) Contains(n *Node) int {
	for i := range *h {
		if n == (*h)[i] {
			return i
		}
	}
	return -1
}

func (h *NodeList) Remove(i int) *Node {
	if i < 0 {
		return nil
	}
	if i >= len(*h) {
		return nil
	}
	node := (*h)[i]
	*h = append((*h)[:i], (*h)[i+1:]...)
	return node
}

func (h *NodeList) WedgeIn(i int, n *Node) {
	if i < 0 {
		return
	}
	if i >= len(*h) {
		*h = append(*h, n)
		return
	}
	*h = append((*h)[:i+1], (*h)[i:]...)
	(*h)[i] = n
}

func (h *NodeList) Pop() *Node {
	if len(*h) == 0 {
		return nil
	}
	popped := (*h)[len((*h))-1]
	*h = (*h)[:len((*h))-1]
	return popped
}

func (h *NodeList) PopUntil(first string, rest ...string) *Node {
	var popped *Node
	for {
		popped = h.Pop()
		if popped == nil {
			return nil
		}

		if popped.NodeName == first {
			return popped
		}
		for _, tagName := range rest {
			if popped.NodeName == tagName {
				return popped
			}
		}
	}
}

func (h *NodeList) PopUntilConditions(funcs ...func(e *Node) bool) *Node {
	for {
		last := len(*h) - 1
		if last < 0 {
			return nil
		}
		for _, f := range funcs {
			if f((*h)[last]) {
				return (*h)[last]
			}
		}

		h.Pop()
	}
}

type StackOfOpenElements struct {
	NodeList
}

func (s *StackOfOpenElements) Push(n *Node) {
	s.NodeList = append(s.NodeList, n)
}

type ActiveFormattingElements struct {
	NodeList
}

func (s *ActiveFormattingElements) Push(n *Node) {
	if len(s.NodeList) < 3 {
		s.NodeList = append(s.NodeList, n)
		return
	}

	iter := NewNodeIterator(s.NodeList)
	rewinder := NewNodeRewinder(s.NodeList)
	// rewind to the last marker or the top of the list
	for rewinder.Prev() {
		node := rewinder.Node()
		if node == ScopeMarker {
			iter.WithStartFrom(node)
			break
		}
	}

	// look through the list of active formatting elements for similar
	// elements (same name, namespace, and value)
	similarNodes := []*Node{}
	for iter.Next() {
		node := iter.Node()
		if !compareNodes(node, n) {
			continue
		}

		// then remove the earliest such element from the list of active formatting elements.
		similarNodes = append(similarNodes, node)
		if len(similarNodes) >= 3 {
			s.NodeList.Remove(s.NodeList.Contains(similarNodes[0]))
			similarNodes = similarNodes[:len(similarNodes)-1]
		}
	}

	s.NodeList = append(s.NodeList, n)
}

func compareNodes(a, b *Node) bool {
	if a.NodeName != b.NodeName {
		return false
	}

	if a.Element.NamespaceURI != b.Element.NamespaceURI {
		return false
	}

	if a.Attributes.Length != b.Attributes.Length {
		return false
	}

	for k, v := range b.Attributes.Attrs {
		e := a.Attributes.GetNamedItem(k)
		if e == nil {
			return false
		}
		if v.Namespace != e.Namespace {
			return false
		}

		if v.Name != e.Name {
			return false
		}

		if v.Value != e.Value {
			return false
		}
	}

	return true
}
