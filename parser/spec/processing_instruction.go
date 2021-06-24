package spec

// https:domspec.whatwg.org/#processinginstruction
type ProcessingInstruction struct {
	Target string
	*CharacterData
}
