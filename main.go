package main

import (
	"fmt"
	"strings"

	"github.com/heathj/gobrowse/parser"
)

func main() {
	p := parser.NewParser(strings.NewReader("&#x0000;"))
	tokens, err := p.Start()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", tokens)
}
