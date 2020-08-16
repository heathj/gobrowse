package dom

import "browser/parser/webidl"

// https://dom.spec.whatwg.org/#characterdata
type CharacterData struct {
	Data   webidl.DOMString
	Length uint

	*NodeFields
}

func (c *CharacterData) substringData(offset, count uint) webidl.DOMString     { return "" }
func (c *CharacterData) appendData(data webidl.DOMString)                      {}
func (c *CharacterData) insertData(offset uint, data webidl.DOMString)         {}
func (c *CharacterData) deleteData(offset, count uint)                         {}
func (c *CharacterData) replaceData(offset, count uint, data webidl.DOMString) {}
