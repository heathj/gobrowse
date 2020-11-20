package parser

import (
	"browser/parser/spec"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"testing"
)

type treeTest struct {
	in       string
	expected string
}

func parseTests(t *testing.T) []treeTest {
	data, err := ioutil.ReadFile("./tests/tree_construction/passing.dat")
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
		t.in = splits[0]
		for i := range splits {
			switch splits[i] {
			case "#errors":
			case "#document":
				t.expected = "#document\n"
				for j := i + 1; j < len(splits); j++ {
					if len(splits[j]) == 0 {
						continue
					}
					if splits[j][0] != '|' {
						continue
					}

					t.expected += splits[j] + "\n"
				}
			}
		}

		treeTests = append(treeTests, t)
	}

	return treeTests
}

func serializeNodeType(node *spec.Node, ident int) string {
	switch node.NodeType {
	case spec.ElementNode:
		e := "<" + string(node.NodeName)
		if node.Attributes != nil && len(node.Attributes.Attrs) != 0 {
			e += ">"
			keys := make([]string, 0, len(node.Attributes.Attrs))
			for name := range node.Attributes.Attrs {
				keys = append(keys, name)
			}
			sort.Strings(keys)
			spaces := "| "
			for i := 1; i < ident; i++ {
				spaces += "  "
			}
			for _, name := range keys {
				value := node.Attributes.Attrs[name]
				e += "\n" + spaces + name + "=\"" + value + "\""
			}
		} else {
			e += ">"
		}
		return e
	case spec.TextNode:
		return "\"" + string(node.Text.Data) + "\""
	case spec.CommentNode:
		return "<!-- " + string(node.Comment.Data) + " -->"
	case spec.DocumentTypeNode:
		d := "<!DOCTYPE " + string(node.DocumentType.Name)
		if len(node.DocumentType.PublicID) == 0 && len(node.DocumentType.SystemID) == 0 {
			return d
		}
		if len(node.DocumentType.PublicID) != 0 && string(node.DocumentType.PublicID) != missing {
			d += " \"" + string(node.DocumentType.PublicID) + "\""
		}
		if len(node.DocumentType.SystemID) != 0 && string(node.DocumentType.SystemID) != missing {
			d += " \"" + string(node.DocumentType.SystemID) + "\""
		}

		d += ">"
		return d
	case spec.DocumentNode:
		return "#document"
	case spec.ProcessingInstructionNode:
		return "<?" + string(node.ProcessingInstruction.CharacterData.Data) + ">"
	default:
		fmt.Printf("Error serializing node: %+v\n", node)
		return ""
	}

}

func serialize(node *spec.Node, ident int) string {
	ser := serializeNodeType(node, ident+1) + "\n"
	if node.NodeType != spec.DocumentNode {
		spaces := "| "
		for i := 1; i < ident; i++ {
			spaces += "  "
		}
		ser = spaces + ser
	}
	for _, child := range node.ChildNodes {
		ser += serialize(child, ident+1)
	}

	return ser
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
		p, tcc, sc, wg := NewHTMLTokenizer(test.in, htmlParserConfig{debug: 0})
		tc := NewHTMLTreeConstructor(tcc, sc, wg)
		wg.Add(3)
		go tc.ConstructTree()
		go p.Tokenize()

		wg.Wait()
		s := serialize(tc.HTMLDocument.Node, 0)

		if s != test.expected {
			t.Errorf("Wrong document. Expected: \n\n%s\nGot: \n\n%s", test.expected, s)
		}
	})
}
