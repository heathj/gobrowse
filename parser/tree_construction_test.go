package parser

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/heathj/gobrowse/parser/spec"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	htmlIn     string
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
			docFrag = spec.NewDOMElement(nil, splits[i+1], spec.Htmlns)
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
	data, err := os.ReadFile("./tests/tree_construction/tests23.dat")
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
			t.htmlIn += s + "\n"
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

		if len(t.htmlIn) > 0 {
			t.htmlIn = t.htmlIn[:len(t.htmlIn)-1]
		}
		t.expected, t.docFrag.context = getExpectedAndDocFrag(splits)
		treeTests = append(treeTests, t)
	}

	return treeTests
}

func TestTreeConstructorAll(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	// used to print tabs and newlines so that we can visualize the trees
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
	})
	tests := parseTests(t)
	for _, test := range tests {
		if test.scriptMode == scriptBoth {
			runTreeConstructorTest(t, test, false)
			//	runTreeConstructorTest(t, test, true)
		} else {
			if test.scriptMode == scriptOn {
				runTreeConstructorTest(t, test, true)
			} else {
				runTreeConstructorTest(t, test, false)
			}
		}
	}

}

func runTreeConstructorTest(t *testing.T, test treeTest, scriptingEnabled bool) {
	t.Run(fmt.Sprintf("%s-scripting-%t", test.htmlIn, scriptingEnabled), func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)

		var err error
		if test.docFrag.enabled {
			err = testParseHTMLFragment(assert, test, scriptingEnabled)
		} else {
			err = testTreeConstructor(assert, test, scriptingEnabled)
		}

		assert.Nil(err)
	})
}

func testParseHTMLFragment(assert *assert.Assertions, test treeTest, scriptingEnabled bool) error {
	nodes := ParseHTMLFragment(test.docFrag.context, test.htmlIn, noQuirks, scriptingEnabled)
	n := spec.NewHTMLDocumentNode()
	for _, node := range nodes {
		n.AppendChild(node)
	}
	assert.Equal(test.expected, n.Node.String(), "these trees should be equal")
	return nil
}

func testTreeConstructor(assert *assert.Assertions, test treeTest, scriptingEnabled bool) error {
	p := NewParser(strings.NewReader(test.htmlIn))
	p.TreeConstructor.scriptingEnabled = scriptingEnabled
	tree, err := p.Start()
	if err != nil {
		return errors.Wrap(err, "error running the parser")
	}

	assert.Equal(test.expected, tree.String(), "these trees should be equal")
	return nil
}
