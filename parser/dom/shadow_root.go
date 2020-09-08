package dom

type ShadowRoot struct {
	mode ShadowRootMode
	host *Element
	//onslotchange EventHandler
	*DocumentFragment
}
