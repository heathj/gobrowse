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

type scriptingMode uint

const (
	scriptBoth scriptingMode = iota
	scriptOff
	scriptOn
)

type treeTest struct {
	in         string
	docFrag    docFramementTest
	scriptMode scriptingMode
	expected   string
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
			} else if s == "#script-on" {
				t.scriptMode = scriptOn
			} else if s == "#script-off" {
				t.scriptMode = scriptOff
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
		if test.scriptMode == scriptBoth {
			runTreeConstructorTest(test, t, false)
			runTreeConstructorTest(test, t, true)
		} else {
			if test.scriptMode == scriptOn {
				runTreeConstructorTest(test, t, true)
			} else {
				runTreeConstructorTest(test, t, false)
			}
		}
	}

}

func runTreeConstructorTest(test treeTest, t *testing.T, scriptingEnabled bool) {
	t.Run(test.in, func(t *testing.T) {
		t.Parallel()
		if test.docFrag.enabled {
			nodes := ParseHTMLFragment(test.docFrag.context, test.in, noQuirks, scriptingEnabled)
			n := spec.NewHTMLDocumentNode()
			for _, node := range nodes {
				n.AppendChild(node)
			}
			s := n.Node.String()
			if s != test.expected {
				t.Errorf("Wrong document. Expected: \n\n%s\nGot: \n\n%s", test.expected, s)
			}
		} else {
			p, tcc, sc, wg := NewHTMLTokenizer(test.in, htmlParserConfig{debug: 0})
			tc := NewHTMLTreeConstructor(tcc, sc, wg)
			tc.scriptingEnabled = scriptingEnabled
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
