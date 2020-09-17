package spec

func NewNamedNodeMap(attrs map[string]string) *NamedNodeMap {
	return &NamedNodeMap{
		Length: len(attrs),
		attrs:  attrs,
	}
}

type NamedNodeMap struct {
	Length int
	attrs  map[string]string
}

func (n *NamedNodeMap) Item(index int) *Attr            { return nil }
func (n *NamedNodeMap) GetNamedItem() *Attr             { return nil }
func (n *NamedNodeMap) GetNamedItemNS() *Attr           { return nil }
func (n *NamedNodeMap) SetNamedItem(attr *Attr) *Attr   { return nil }
func (n *NamedNodeMap) SetNamedItemNS(attr *Attr) *Attr { return nil }
func (n *NamedNodeMap) RemoveNamedItem() *Attr          { return nil }
func (n *NamedNodeMap) RemoveNamedItemNS() *Attr        { return nil }
