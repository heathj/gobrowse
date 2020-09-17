package spec

type ShadowRootMode string

const (
	open   ShadowRootMode = "open"
	closed ShadowRootMode = "closed"
)

type ShadowRootInit struct {
	mode           ShadowRootMode
	delegatesFocus bool
}
