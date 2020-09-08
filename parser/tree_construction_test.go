package parser

import (
	"browser/parser/dom"
	"testing"
)

// compareDocuments returns true if the two documents are the same and false otherwise.
func compareDocuments(a, b *dom.Node) bool {
	if len(a.ChildNodes) != len(b.ChildNodes) {
		return false
	}

	if a.Doctype.Name != b.Doctype.Name {
		return false
	}

	for i := 0; i < len(a.ChildNodes); i++ {
		if !compareNodes(a.ChildNodes[i], b.ChildNodes[i]) {
			return false
		}
	}

	return true
}

func compareNodes(a, b *dom.Node) bool {
	if !a.IsEqualNode(b) {
		return false
	}

	ret := false
	if len(a.ChildNodes) > 0 {
		ret = compareNodes(a.ChildNodes[0], b.ChildNodes[0])
	}

	if !ret {
		return false
	}

	if a.NextSibling == nil && b.NextSibling == nil {
		return true
	}
	if a.NextSibling == nil && b.NextSibling == nil {
		return false
	}

	return compareNodes(a.NextSibling, b.NextSibling)
}

type documentBuilderAccuracyTestCase struct {
	inHTML      string
	outDocument *dom.Node
}

var documentBuilderAccuracyTests = []documentBuilderAccuracyTestCase{
	{`<DOCTYPE html>
	<html>
		<head></head>
		<body>
			<b>JAKE</b>
		</body>
	</html>`, makeBTestCase()},
}

func makeBTestCase() *dom.Node {
	text := &dom.Node{
		NodeType:  dom.TextNode,
		NodeValue: "JAKE",
	}
	b := &dom.Node{
		NodeType:   dom.ElementNode,
		NodeName:   "B",
		ChildNodes: dom.NodeList{text},
		FirstChild: text,
		LastChild:  text,
	}
	body := &dom.Node{
		NodeType:   dom.ElementNode,
		NodeName:   "BODY",
		ChildNodes: dom.NodeList{b},
		FirstChild: b,
		LastChild:  b,
	}
	head := &dom.Node{
		NodeType:    dom.ElementNode,
		NodeName:    "HEAD",
		NextSibling: body,
	}
	doc := &dom.Node{
		NodeType:    dom.DocumentTypeNode,
		NodeName:    "DOCTYPE",
		NextSibling: body,
	}
	dom := &dom.Node{
		ChildNodes: dom.NodeList{
			doc, head, body,
		},
	}

	return dom
}

func TestDocumentBuildAccuracy(t *testing.T) {
	for _, tt := range documentBuilderAccuracyTests {
		runDocumentBuilderAccuracyTest(tt, t)
	}
}

func runDocumentBuilderAccuracyTest(tt documentBuilderAccuracyTestCase, t *testing.T) {
	t.Run(tt.inHTML, func(t *testing.T) {
		t.Parallel()
		p, tc, wg := NewHTMLTokenizer(tt.inHTML, nil)
		trc := NewHTMLTreeConstructor(tc, wg)
		go p.Tokenize()
		go trc.ConstructTree()

		wg.Wait()
		if !compareDocuments(trc.HTMLDocument.Node, tt.outDocument) {
			t.Errorf("Tree constructor built a different document than expected. Got %s, expected %s", trc.HTMLDocument.Node, tt.outDocument)
		}
	})
}
