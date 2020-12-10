package parser

import (
	"browser/parser/spec"
	"browser/parser/webidl"
	"io/ioutil"
	"strings"
	"testing"
)

type docFramementTest struct {
	enabled bool
	context *spec.Node
}

type treeTest struct {
	in       string
	docFrag  docFramementTest
	expected string
}

func getExpectedAndDocFrag(splits []string) (string, *spec.Node) {
	expected := ""
	var docFrag *spec.Node
	for i := range splits {
		switch splits[i] {
		case "#errors":
		case "#document-fragment":
			docFrag = spec.NewDOMElement(nil, webidl.DOMString(splits[i+1]), "html")
		case "#document":
			expected = "#document\n"
			for j := i + 1; j < len(splits); j++ {
				if len(splits[j]) == 0 {
					continue
				}

				expected += splits[j] + "\n"
			}
			return expected, docFrag
		}
	}
	return expected, docFrag
}

func parseTests(t *testing.T) []treeTest {
	data, err := ioutil.ReadFile("./tests/tree_construction/basic.dat")
	if err != nil {
		t.Error(err)
		return nil
	}

	tests := strings.Split(string(data), "#data\n")
	var treeTests []treeTest
	for i, test := range tests {
		if i == 0 {
			continue
		}
		t := treeTest{}
		splits := strings.Split(test, "\n")
		for _, s := range splits {
			if s == "#document" || s == "#errors" {
				break
			}
			t.in += s + "\n"
		}
		for _, s := range splits {
			if s == "#document-fragment" {
				t.docFrag.enabled = true
			}
		}

		if len(t.in) > 0 {
			t.in = t.in[:len(t.in)-1]
		}
		t.expected, t.docFrag.context = getExpectedAndDocFrag(splits)
		treeTests = append(treeTests, t)
	}

	return treeTests
}

func TestTreeConstructor(t *testing.T) {
	tests := parseTests(t)
	for _, test := range tests {
		runTreeConstructorTest(test, t)
	}

}

func runTreeConstructorTest(test treeTest, t *testing.T) {
	t.Run(test.in, func(t *testing.T) {
		t.Parallel()
		if test.docFrag.enabled {
			nodes := ParseHTMLFragment(test.docFrag.context, test.in, noQuirks, false)
			n := spec.NewDOMElement(nil, "html", "html")
			for _, node := range nodes {
				n.AppendChild(node)
			}
			s := n.String()

			if s != test.expected {
				t.Errorf("Wrong document. Expected: \n\n%s\nGot: \n\n%s", test.expected, s)
			}
		} else {
			p, tcc, sc, wg := NewHTMLTokenizer(test.in, htmlParserConfig{debug: 0})
			tc := NewHTMLTreeConstructor(tcc, sc, wg)
			wg.Add(3)
			go tc.ConstructTree()
			go p.Tokenize()

			wg.Wait()
			s := tc.HTMLDocument.Node.String()

			if s != test.expected {
				t.Errorf("Wrong document. Expected: \n\n%s\nGot: \n\n%s", test.expected, s)
			}
		}
	})
}
