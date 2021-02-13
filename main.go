package main

import "github.com/heathj/gobrowse/parser"

func main() {
	p, _, _, _ := parser.NewHTMLTokenizer("<html><head></head><body></body>", nil)
	p.Tokenize()
}
