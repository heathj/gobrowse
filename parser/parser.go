package parser

import (
	"io"

	"github.com/heathj/gobrowse/parser/spec"
)

type Parser struct {
	Tokenizer       *HTMLTokenizer
	TreeConstructor *HTMLTreeConstructor
}

func NewParser(htmlIn io.Reader) *Parser {
	tokenizer := NewHTMLTokenizer(htmlIn)
	treeConstructor := NewHTMLTreeConstructor()
	return &Parser{
		Tokenizer:       tokenizer,
		TreeConstructor: treeConstructor,
	}
}

type Progress struct {
	AdjustedCurrentNode *spec.Node
	TokenizerState      *tokenizerState
}

func MakeProgress(adjCurNode *spec.Node, tokenizerState *tokenizerState) *Progress {
	return &Progress{
		AdjustedCurrentNode: adjCurNode,
		TokenizerState:      tokenizerState,
	}
}

func (p *Parser) Start() (*spec.Node, error) {
	start := dataState
	if err := p.startAt(&start); err != nil {
		return nil, err
	}
	return p.TreeConstructor.HTMLDocument.Node, nil
}

// start parsing the tokens at a specific start point
func (p *Parser) startAt(startState *tokenizerState) error {
	progress := MakeProgress(nil, startState)
	for p.Tokenizer.Next() {
		t, err := p.Tokenizer.Token(progress)
		if err != nil {
			return err
		}
		progress = p.TreeConstructor.ProcessToken(*t)
	}

	return nil
}

// startAtTokens returns the set of tokens that were produced from this input.
// mainly used for testing and debugging tokenizer.
func (p *Parser) startAtTokens(startState *tokenizerState) ([]Token, error) {
	var (
		progress *Progress = MakeProgress(nil, startState)
		tokens             = []Token{}
	)
	for p.Tokenizer.Next() {
		t, err := p.Tokenizer.Token(progress)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, *t)
		progress = p.TreeConstructor.ProcessToken(*t)
	}

	return tokens, nil
}
