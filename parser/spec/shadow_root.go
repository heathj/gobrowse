package spec

type ShadowRoot struct {
	mode ShadowRootMode
	host *Element
	//onslotchange EventHandler
	*DocumentFragment
}
