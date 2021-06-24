package spec

// https:domspec.whatwg.org/#text
type Text struct {
	wholeText string
	*CharacterData
}

func NewText(data string) *Text {
	return &Text{
		wholeText: data,
		CharacterData: &CharacterData{
			Data:   data,
			Length: len(data),
		}}
}

func (t *Text) splitText(offset uint) *Text {
	return nil
}
