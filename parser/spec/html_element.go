package spec

func NewHTMLElement(name string) *HTMLElement {
	elem := &HTMLElement{}
	switch name {
	case "script":
		elem.HTMLScript = &HTMLScript{}
	case "document":
		elem.HTMLDocument = &HTMLDocument{}
	case "form":
		elem.HTMLForm = &HTMLForm{}
	case "head":
		elem.HTMLHead = &HTMLHead{}
	case "location":
		elem.HTMLLocation = &HTMLLocation{}
	case "table":
		elem.HTMLTable = &HTMLTable{}
	case "tbody":
		elem.HTMLTBody = &HTMLTBody{}
	case "template":
		elem.HTMLTemplate = &HTMLTemplate{}
	case "tfoot":
		elem.HTMLTFoot = &HTMLTFoot{}
	case "thead":
		elem.HTMLTHead = &HTMLTHead{}
	case "tr":
		elem.HTMLTr = &HTMLTr{}
	}

	return elem
}

type HTMLElement struct {
	Title, Lang, Dir, AccessKey, AccessKeyLabel, Autocapitalize, InnerText string
	Translate, Hidden, Draggable, Spellcheck                               bool

	*HTMLScript
	*HTMLDocument
	*HTMLForm
	*HTMLHead
	*HTMLLocation
	*HTMLTable
	*HTMLTBody
	*HTMLTemplate
	*HTMLTFoot
	*HTMLTHead
	*HTMLTr
	*HTMLWindow
}

func (e *HTMLElement) attachInternals() *ElementInternals { return nil }
func (e *HTMLElement) click()                             {}
