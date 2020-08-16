package dom

// https://dom.spec.whatwg.org/#characterdata
type CharacterData struct {
	data   DOMString
	length uint

	*NodeFields
}

func (c *CharacterData) substringData(offset, count uint) DOMString     { return "" }
func (c *CharacterData) appendData(data DOMString)                      {}
func (c *CharacterData) insertData(offset uint, data DOMString)         {}
func (c *CharacterData) deleteData(offset, count uint)                  {}
func (c *CharacterData) replaceData(offset, count uint, data DOMString) {}
