package spec

// CharacterData is https:domspec.whatwg.org/#characterdata
type CharacterData struct {
	Data   string
	Length int
}

func (c *CharacterData) substringData(offset, count uint) string     { return "" }
func (c *CharacterData) appendData(data string)                      {}
func (c *CharacterData) insertData(offset uint, data string)         {}
func (c *CharacterData) deleteData(offset, count uint)               {}
func (c *CharacterData) replaceData(offset, count uint, data string) {}
