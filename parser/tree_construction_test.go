package parser

import (
	"io/ioutil"
	"strings"
	"testing"
)

type treeTest struct {
	in       string
	expected string
}

func getExpected(splits []string) string {
	ret := ""
	for i := range splits {
		switch splits[i] {
		case "#errors":
		case "#document":
			ret = "#document\n"
			for j := i + 1; j < len(splits); j++ {
				if len(splits[j]) == 0 {
					continue
				}
				if len(splits[j]) == 0 {
					continue
				}

				ret += splits[j] + "\n"
			}
			return ret
		}
	}
	return ret
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
		for i := 0; i < len(splits); i++ {
			if splits[i] == "#document" || splits[i] == "#errors" {
				break
			}
			t.in += splits[i] + "\n"
		}
		if len(t.in) > 0 {
			t.in = t.in[:len(t.in)-1]
		}
		t.expected = getExpected(splits)
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
	})
}
