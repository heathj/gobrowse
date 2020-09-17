package parser

import (
	"browser/parser/spec"
	"testing"
)

// compareDocuments returns true if the two documents are the same and false otherwise.
func compareDocuments(a, b *spec.Node) bool {
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

func compareNodes(a, b *spec.Node) bool {
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
	outDocument *spec.Node
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

func makeBTestCase() *spec.Node {
	text := &spec.Node{
		NodeType:  spec.TextNode,
		NodeValue: "JAKE",
	}
	b := &spec.Node{
		NodeType:   spec.ElementNode,
		NodeName:   "B",
		ChildNodes: spec.NodeList{text},
		FirstChild: text,
		LastChild:  text,
	}
	body := &spec.Node{
		NodeType:   spec.ElementNode,
		NodeName:   "BODY",
		ChildNodes: spec.NodeList{b},
		FirstChild: b,
		LastChild:  b,
	}
	head := &spec.Node{
		NodeType:    spec.ElementNode,
		NodeName:    "HEAD",
		NextSibling: body,
	}
	doc := &spec.Node{
		NodeType:    spec.DocumentTypeNode,
		NodeName:    "DOCTYPE",
		NextSibling: body,
	}
	dom := &spec.Node{
		ChildNodes: spec.NodeList{
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
