package spec

import "browser/parser/webidl"

// CharacterData is https:domspec.whatwg.org/#characterdata
type CharacterData struct {
	Data   webidl.DOMString
	Length int
}

func (c *CharacterData) substringData(offset, count uint) webidl.DOMString     { return "" }
func (c *CharacterData) appendData(data webidl.DOMString)                      {}
func (c *CharacterData) insertData(offset uint, data webidl.DOMString)         {}
func (c *CharacterData) deleteData(offset, count uint)                         {}
func (c *CharacterData) replaceData(offset, count uint, data webidl.DOMString) {}
