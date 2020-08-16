package dom

type howRange uint

const (
	starToStart howRange = iota
	starToEnd
	endToEnd
	endToStart
)

//https://dom.spec.whatwg.org/#range
type Range struct {
	commonAncestorContainer *Node

	AbstractRange
}

func (r *Range) setStart(node *Node, offset uint)                            {}
func (r *Range) setEnd(node *Node, offset uint)                              {}
func (r *Range) setStartBefore(node *Node)                                   {}
func (r *Range) setStartAfter(node *Node)                                    {}
func (r *Range) setEndBefore(node *Node)                                     {}
func (r *Range) setEndAfter(node *Node)                                      {}
func (r *Range) collapse(toStart bool)                                       {}
func (r *Range) collapseDef()                                                {}
func (r *Range) selectNode(node *Node)                                       {}
func (r *Range) selectNodeContents(node *Node)                               {}
func (r *Range) compareBoundaryPoints(how howRange, sourceRange Range) int16 { return 0 }
func (r *Range) deleteContents()                                             {}
func (r *Range) extractContents() *DocumentFragment                          { return nil }
func (r *Range) cloneContents() *DocumentFragment                            { return nil }
func (r *Range) insertNode(node *Node)                                       {}
func (r *Range) surroundContents(newparent *Node)                            {}
func (r *Range) cloneRange() *Range                                          { return nil }
func (r *Range) detach()                                                     {}
func (r *Range) isPointInRange(node *Node, offset uint) bool                 { return false }
func (r *Range) comparePoint(node *Node, offset uint) int16                  { return 0 }
func (r *Range) intersectsNode(node *Node) bool                              { return false }
