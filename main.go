package main

import "browser/parser"

func main() {
	p, _, _, _ := parser.NewHTMLTokenizer("<html><head></head><body></body>", nil)
	p.Tokenize()
}
