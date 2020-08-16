package dom

//https://dom.spec.whatwg.org/#interface-comment
type Comment struct {
	CharacterData

	*NodeFields
}
