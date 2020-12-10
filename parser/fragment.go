package parser

import "browser/parser/spec"

func SerializeHTMLFrame(fragment *spec.Node) string {
	ret := ""
	switch fragment.NodeName {
	case "basefont", "bgsound", "frame", "keygen":
		return ret
	}

	for _, child := range fragment.ChildNodes {
		switch child.NodeType {
		case spec.ElementNode:
			ret += "<" + string(child.NodeName)
			if child.Element.is
		case spec.TextNode:
		case spec.CommentNode:
		case spec.ProcessingInstructionNode:
		case spec.DocumentTypeNode:

		}
	}
	return ret
}

func ParseHTMLFragment(context *spec.Node, input string, quirks quirksMode, scriptingEnabled bool) []*spec.Node {
	tokenizer, tokChan, stateChan, wg := NewHTMLTokenizer(input, htmlParserConfig{})
	treeConstructor := NewHTMLTreeConstructor(tokChan, stateChan, wg)

	treeConstructor.quirksMode = quirks
	treeConstructor.createdBy = htmlFragmentParsingAlgorithm
	var startState tokenizerState
	switch context.NodeName {
	case "title", "textarea":
		startState = rcDataState
	case "style", "xmp", "iframe", "noembed", "noframes":
		startState = rawTextState
	case "script":
		startState = scriptDataState
	case "noscript":
		if scriptingEnabled {
			startState = rawTextState
		} else {
			startState = dataState
		}
	case "plaintext":
		startState = plaintextState
	default:
		startState = dataState
	}
	wg.Add(3)
	go tokenizer.tokenizeStartState(startState)

	n := spec.NewDOMElement(treeConstructor.HTMLDocument.Node, "html", "html")
	n.OwnerDocument = treeConstructor.HTMLDocument.Node
	treeConstructor.HTMLDocument.AppendChild(n)
	treeConstructor.stackOfOpenElements.Push(n)

	if context.NodeName == "template" {
		treeConstructor.stackOfTemplateInsertionModes[0] = inTemplate
	}

	/*tok := &Token{
		TokenType:  startTagToken,
		TagName:    string(context.NodeName),
		Attributes: context.Attributes.Attrs,
	}*/
	im := treeConstructor.resetInsertionMode()
	var next *spec.Node = context.ParentNode
	for {
		if next == nil {
			break
		}
		if next.NodeName == "form" {
			treeConstructor.formElementPointer = next
			break
		}

		next = next.ParentNode
	}

	go treeConstructor.constructTreeStartState(im)
	wg.Wait()

	return n.ChildNodes
}
