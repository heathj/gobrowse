package spec

type ParentNode struct {
	children                            HTMLCollection
	firstElementChild, lastElementChild Element
	childElementCount                   uint16
}

func (p *ParentNode) Prepend()          {}
func (p *ParentNode) Append()           {}
func (p *ParentNode) ReplaceChildren()  {}
func (p *ParentNode) QuerySelector()    {}
func (p *ParentNode) QuerySelectorAll() {}
