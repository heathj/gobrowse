package parser

import (
	"browser/parser/spec"
	"sort"
	"strings"
)

// https://html.spec.whatwg.org/#escapingString
func escapeString(s string, attrVal bool) string {
	s = strings.Replace(s, "&", "&amp;", -1)
	s = strings.Replace(s, "\u00A0", "&nbsp;", -1)
	if attrVal {
		s = strings.Replace(s, "\"", "&quot;", -1)
	} else {
		s = strings.Replace(s, "<", "&lt;", -1)
		s = strings.Replace(s, ">", "&gt;", -1)
	}

	return s
}

func SerializeHTMLFragement(fragment *spec.Node) string {
	ret := ""
	switch fragment.NodeName {
	case "basefont", "bgsound", "frame", "keygen":
		return ret
	}

	for _, child := range fragment.ChildNodes {
		switch child.NodeType {
		case spec.ElementNode:
			ret += "<" + string(child.NodeName)

			// TODO: implementation defined, but needs to be stable
			keys := make([]string, 0, len(child.Attributes.Attrs))
			for name := range child.Attributes.Attrs {
				keys = append(keys, name)
			}
			sort.Strings(keys)
			for _, k := range keys {
				ret += " " + k + "=" + "\"" + escapeString(child.Attributes.Attrs[k], true) + "\""
			}
			ret += ">"
			ret += SerializeHTMLFragement(child) + "</" + string(child.NodeName) + ">"
		case spec.TextNode:
			switch child.ParentNode.NodeName {
			case "style", "script", "xmp", "iframe", "noembed", "noframes", "plaintext":
				ret += string(child.Text.Data)
			default:
				// TODO: and scripting enabled
				if child.ParentNode.NodeName == "noscript" {
					ret += string(child.Text.Data)
				} else {
					ret += escapeString(string(child.Text.Data), false)
				}
			}
		case spec.CommentNode:
			ret += "<!--" +
				string(child.Comment.Data) +
				"-->"
		case spec.ProcessingInstructionNode:
			ret += "<?" +
				string(child.ProcessingInstruction.Target) +
				" " +
				string(child.ProcessingInstruction.Data) +
				">"
		case spec.DocumentTypeNode:
			ret += "<!DOCTYPE" +
				" " +
				string(child.DocumentType.Name) +
				">"
		}
	}
	return ret
}

func ParseHTMLFragment(context *spec.Node, input string, quirks quirksMode, scriptingEnabled bool) []*spec.Node {
	tokenizer, tokChan, stateChan, wg := NewHTMLTokenizer(input, htmlParserConfig{})
	treeConstructor := NewHTMLTreeConstructor(tokChan, stateChan, wg)

	treeConstructor.context = context
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

	n := spec.NewDOMElement(treeConstructor.HTMLDocument.Node, "html", spec.Htmlns)
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
	im := treeConstructor.resetInsertionModeWithContext(context)
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
