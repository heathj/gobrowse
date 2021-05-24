package parser

import (
	"sort"
	"strings"

	"github.com/heathj/gobrowse/parser/webidl"

	"github.com/heathj/gobrowse/parser/spec"
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
				keys = append(keys, string(name))
			}
			sort.Strings(keys)
			for _, k := range keys {
				ret += " " + k + "=" + "\"" + escapeString(string(child.Attributes.Attrs[webidl.DOMString(k)].Value), true) + "\""
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
	parser := NewParser(strings.NewReader(input))
	parser.TreeConstructor.context = context
	parser.TreeConstructor.quirksMode = quirks
	parser.TreeConstructor.createdBy = htmlFragmentParsingAlgorithm
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

	n := spec.NewDOMElement(parser.TreeConstructor.HTMLDocument.Node, "html", spec.Htmlns)
	n.OwnerDocument = parser.TreeConstructor.HTMLDocument.Node
	parser.TreeConstructor.HTMLDocument.AppendChild(n)
	parser.TreeConstructor.stackOfOpenElements.Push(n)

	if context.NodeName == "template" {
		parser.TreeConstructor.stackOfTemplateInsertionModes[0] = inTemplate
	}

	parser.TreeConstructor.curInsertionMode = parser.TreeConstructor.resetInsertionModeWithContext(context)
	var next *spec.Node = context.ParentNode
	for {
		if next == nil {
			break
		}
		if next.NodeName == "form" {
			parser.TreeConstructor.formElementPointer = next
			break
		}

		next = next.ParentNode
	}

	parser.startAt(&startState)
	return n.ChildNodes
}
