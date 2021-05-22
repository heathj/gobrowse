package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type tokezinerAttributeAccuracyTestcase struct {
	inHTML string            // snippet of HTML to tokenize (should only be one element)
	attrs  map[string]string // expected attributes to collected from the first token that is produced
}

var tokenizerAttributeAccuracyTests = []tokezinerAttributeAccuracyTestcase{
	{"<head></head>", map[string]string{}},
	{"<script src='123' onload='test'></script>", map[string]string{
		"src":    "123",
		"onload": "test",
	}},
	{"<a href='https://google.com' onclick='alert(1)'>Click this</a>", map[string]string{
		"href":    "https://google.com",
		"onclick": "alert(1)",
	}},
	{"<script src='123' src='456'></script>", map[string]string{
		"src": "123",
	}},
	{"<script src=123 onload=test></script>", map[string]string{
		"src":    "123",
		"onload": "test",
	}},
	{"<script src='123' onload='test' ></script>", map[string]string{
		"src":    "123",
		"onload": "test",
	}},
	{"<script =src='123'onload='test' ></script>", map[string]string{
		"=src":   "123",
		"onload": "test",
	}},
	{"<script src></script>", map[string]string{
		"src": "",
	}},
	{"<script src test></script>", map[string]string{
		"src":  "",
		"test": "",
	}},
	{"<script 'asd></script>", map[string]string{
		"'asd": "",
	}},
	{"<script <asd></script>", map[string]string{
		"<asd": "",
	}},
	{"<script ABC=123></script>", map[string]string{
		"abc": "123",
	}},
	{"<script abc='\u0000123'></script>", map[string]string{
		"abc": "\uFFFD123",
	}},
	{"<script abc=></script>", map[string]string{
		"abc": "",
	}},
	{"<script\tabc=123></script>", map[string]string{
		"abc": "123",
	}},
}

// TestTokenizerAttributeAccuracy just makes sure that we have the
// correct number attribute names and values
func TestTokenizerAttributeAccuracy(t *testing.T) {
	for _, tt := range tokenizerAttributeAccuracyTests {
		runTestTokenizerAttributeAccuracy(tt, t)
	}
}

// helper function to parallelize the above test case.
func runTestTokenizerAttributeAccuracy(tt tokezinerAttributeAccuracyTestcase, t *testing.T) {
	t.Run(tt.inHTML, func(t *testing.T) {
		t.Parallel()
		p := NewHTMLTokenizer(strings.NewReader(tt.inHTML))
		startState := dataState
		progress := MakeProgress(nil, &startState)
		for p.Next() {
			token, err := p.Token(progress)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range tt.attrs {
				if _, ok := token.Attributes[k]; !ok {
					t.Errorf("Expected to find a key of %s, didn't find one\n", k)
				} else {
					if v != string(token.Attributes[k].Value) {
						t.Errorf("Expected %s as the value, got %s\n", v, token.Attributes[k].Value)
					}
				}
			}
		}
	})
}

type stateMachineTestCase struct {
	inRune            rune           // the rune to pass to the startingState
	startingState     tokenizerState // the state to start from
	shouldReconsume   bool           // the expectation if the next state should reconsume
	nextExpectedState tokenizerState // the next state
}

// TestStateParsers tests to make sure that each component of the state machine returns the next
// expected state. Overall, it checks the basic state machine flows are accurage. Obviously, all
// cases won't be covered because some flows require state, but the basic cases are covered here.
func TestStateParsers(t *testing.T) {
	stateParserTests := []stateMachineTestCase{
		{'&', dataState, false, characterReferenceState},
		{'<', dataState, false, tagOpenState},
		{'\u0000', dataState, false, dataState},
		{'a', dataState, false, dataState},
		{'A', dataState, false, dataState},
		{'z', dataState, false, dataState},
		{'1', dataState, false, dataState},

		{'&', rcDataState, false, characterReferenceState},
		{'<', rcDataState, false, rcDataLessThanSignState},
		{'\u0000', rcDataState, false, rcDataState},
		{'a', rcDataState, false, rcDataState},
		{'#', rcDataState, false, rcDataState},

		{'<', rawTextState, false, rawTextLessThanSignState},
		{'\u0000', rawTextState, false, rawTextState},
		{'a', rawTextState, false, rawTextState},
		{'#', rawTextState, false, rawTextState},

		{'<', scriptDataState, false, scriptDataLessThanSignState},
		{'\u0000', scriptDataState, false, scriptDataState},
		{'a', scriptDataState, false, scriptDataState},
		{'#', scriptDataState, false, scriptDataState},

		{'\u0000', plaintextState, false, plaintextState},
		{'!', plaintextState, false, plaintextState},
		{'2', plaintextState, false, plaintextState},
		{'a', plaintextState, false, plaintextState},

		{'!', tagOpenState, false, markupDeclarationOpenState},
		{'/', tagOpenState, false, endTagOpenState},
		{'a', tagOpenState, true, tagNameState},
		{'A', tagOpenState, true, tagNameState},
		{'z', tagOpenState, true, tagNameState},
		{'?', tagOpenState, true, bogusCommentState},
		{'1', tagOpenState, true, dataState},
		{'2', tagOpenState, true, dataState},

		{'a', endTagOpenState, true, tagNameState},
		{'b', endTagOpenState, true, tagNameState},
		{'A', endTagOpenState, true, tagNameState},
		{'B', endTagOpenState, true, tagNameState},
		{'>', endTagOpenState, false, dataState},
		{'"', endTagOpenState, true, bogusCommentState},
		{'\\', endTagOpenState, true, bogusCommentState},
		{'#', endTagOpenState, true, bogusCommentState},
		{'$', endTagOpenState, true, bogusCommentState},

		{'\t', tagNameState, false, beforeAttributeNameState},
		{'\u000A', tagNameState, false, beforeAttributeNameState},
		{'\u000C', tagNameState, false, beforeAttributeNameState},
		{' ', tagNameState, false, beforeAttributeNameState},
		{'/', tagNameState, false, selfClosingStartTagState},
		{'>', tagNameState, false, dataState},
		{'a', tagNameState, false, tagNameState},
		{'A', tagNameState, false, tagNameState},
		{'\u0000', tagNameState, false, tagNameState},
		{'1', tagNameState, false, tagNameState},

		{'/', rcDataLessThanSignState, false, rcDataEndTagOpenState},
		{'a', rcDataLessThanSignState, true, rcDataState},
		{'b', rcDataLessThanSignState, true, rcDataState},

		{'a', rcDataEndTagOpenState, true, rcDataEndTagNameState},
		{'b', rcDataEndTagOpenState, true, rcDataEndTagNameState},
		{'A', rcDataEndTagOpenState, true, rcDataEndTagNameState},
		{'1', rcDataEndTagOpenState, true, rcDataState},
		{'#', rcDataEndTagOpenState, true, rcDataState},

		// TODO: rcdata end tag name state tests will need some setup ahead of time because it checks
		// if the current end token is valid
		{'A', rcDataEndTagNameState, false, rcDataEndTagNameState},
		{'Z', rcDataEndTagNameState, false, rcDataEndTagNameState},
		{'a', rcDataEndTagNameState, false, rcDataEndTagNameState},
		{'z', rcDataEndTagNameState, false, rcDataEndTagNameState},
		{'1', rcDataEndTagNameState, true, rcDataState},
		{'#', rcDataEndTagNameState, true, rcDataState},

		{'/', rawTextLessThanSignState, false, rawTextEndTagOpenState},
		{'a', rawTextLessThanSignState, true, rawTextState},
		{'1', rawTextLessThanSignState, true, rawTextState},

		{'a', rawTextEndTagOpenState, true, rawTextEndTagNameState},
		{'Z', rawTextEndTagOpenState, true, rawTextEndTagNameState},
		{'z', rawTextEndTagOpenState, true, rawTextEndTagNameState},
		{'1', rawTextEndTagOpenState, true, rawTextState},
		{'@', rawTextEndTagOpenState, true, rawTextState},

		// TODO: rawtext end tag name state tests will need some setup ahead of time because it checks
		// if the current end token is valid
		{'A', rawTextEndTagNameState, false, rawTextEndTagNameState},
		{'Z', rawTextEndTagNameState, false, rawTextEndTagNameState},
		{'a', rawTextEndTagNameState, false, rawTextEndTagNameState},
		{'z', rawTextEndTagNameState, false, rawTextEndTagNameState},
		{'1', rawTextEndTagNameState, true, rawTextState},
		{'#', rawTextEndTagNameState, true, rawTextState},

		{'/', scriptDataLessThanSignState, false, scriptDataEndTagOpenState},
		{'!', scriptDataLessThanSignState, false, scriptDataEscapeStartState},
		{'a', scriptDataLessThanSignState, true, scriptDataState},
		{'@', scriptDataLessThanSignState, true, scriptDataState},

		{'a', scriptDataEndTagOpenState, true, scriptDataEndTagNameState},
		{'Z', scriptDataEndTagOpenState, true, scriptDataEndTagNameState},
		{'z', scriptDataEndTagOpenState, true, scriptDataEndTagNameState},
		{'1', scriptDataEndTagOpenState, true, scriptDataState},
		{'#', scriptDataEndTagOpenState, true, scriptDataState},
		{'$', scriptDataEndTagOpenState, true, scriptDataState},

		// TODO: script data end tag name state tests will need some setup ahead of time because it
		// checks if the current end token is valid
		{'A', scriptDataEndTagNameState, false, scriptDataEndTagNameState},
		{'Z', scriptDataEndTagNameState, false, scriptDataEndTagNameState},
		{'a', scriptDataEndTagNameState, false, scriptDataEndTagNameState},
		{'#', scriptDataEndTagNameState, true, scriptDataState},
		{'^', scriptDataEndTagNameState, true, scriptDataState},

		{'-', scriptDataEscapeStartState, false, scriptDataEscapeStartDashState},
		{'a', scriptDataEscapeStartState, true, scriptDataState},
		{'b', scriptDataEscapeStartState, true, scriptDataState},
		{'@', scriptDataEscapeStartState, true, scriptDataState},

		{'-', scriptDataEscapeStartDashState, false, scriptDataEscapedDashDashState},
		{'a', scriptDataEscapeStartDashState, true, scriptDataState},
		{'@', scriptDataEscapeStartDashState, true, scriptDataState},

		{'-', scriptDataEscapedState, false, scriptDataEscapedDashState},
		{'<', scriptDataEscapedState, false, scriptDataEscapedLessThanSignState},
		{'\u0000', scriptDataEscapedState, false, scriptDataEscapedState},
		{'a', scriptDataEscapedState, false, scriptDataEscapedState},
		{'1', scriptDataEscapedState, false, scriptDataEscapedState},

		{'-', scriptDataEscapedDashState, false, scriptDataEscapedDashDashState},
		{'<', scriptDataEscapedDashState, false, scriptDataEscapedLessThanSignState},
		{'\u0000', scriptDataEscapedDashState, false, scriptDataEscapedState},
		{'a', scriptDataEscapedDashState, false, scriptDataEscapedState},
		{'1', scriptDataEscapedDashState, false, scriptDataEscapedState},

		{'-', scriptDataEscapedDashDashState, false, scriptDataEscapedDashDashState},
		{'<', scriptDataEscapedDashDashState, false, scriptDataEscapedLessThanSignState},
		{'>', scriptDataEscapedDashDashState, false, scriptDataState},
		{'\u0000', scriptDataEscapedDashDashState, false, scriptDataEscapedState},
		{'a', scriptDataEscapedDashDashState, false, scriptDataEscapedState},
		{'$', scriptDataEscapedDashDashState, false, scriptDataEscapedState},

		{'/', scriptDataEscapedLessThanSignState, false, scriptDataEscapedEndTagOpenState},
		{'a', scriptDataEscapedLessThanSignState, true, scriptDataDoubleEscapeStartState},
		{'A', scriptDataEscapedLessThanSignState, true, scriptDataDoubleEscapeStartState},
		{'z', scriptDataEscapedLessThanSignState, true, scriptDataDoubleEscapeStartState},
		{'#', scriptDataEscapedLessThanSignState, true, scriptDataEscapedState},

		{'a', scriptDataEscapedEndTagOpenState, true, scriptDataEscapedEndTagNameState},
		{'B', scriptDataEscapedEndTagOpenState, true, scriptDataEscapedEndTagNameState},
		{'#', scriptDataEscapedEndTagOpenState, true, scriptDataEscapedState},

		// TODO: script data escaped end tag name state tests will need some setup ahead of time
		// because it checks if the current end token is valid
		{'A', scriptDataEscapedEndTagNameState, false, scriptDataEscapedEndTagNameState},
		{'B', scriptDataEscapedEndTagNameState, false, scriptDataEscapedEndTagNameState},
		{'a', scriptDataEscapedEndTagNameState, false, scriptDataEscapedEndTagNameState},
		{'b', scriptDataEscapedEndTagNameState, false, scriptDataEscapedEndTagNameState},
		{'$', scriptDataEscapedEndTagNameState, true, scriptDataEscapedState},
		{'%', scriptDataEscapedEndTagNameState, true, scriptDataEscapedState},

		//TODO: some stateful behavior with the temp. buffer
		{'A', scriptDataDoubleEscapeStartState, false, scriptDataDoubleEscapeStartState},
		{'B', scriptDataDoubleEscapeStartState, false, scriptDataDoubleEscapeStartState},
		{'a', scriptDataDoubleEscapeStartState, false, scriptDataDoubleEscapeStartState},
		{'B', scriptDataDoubleEscapeStartState, false, scriptDataDoubleEscapeStartState},
		{'1', scriptDataDoubleEscapeStartState, true, scriptDataEscapedState},
		{'#', scriptDataDoubleEscapeStartState, true, scriptDataEscapedState},

		{'-', scriptDataDoubleEscapedState, false, scriptDataDoubleEscapedDashState},
		{'<', scriptDataDoubleEscapedState, false, scriptDataDoubleEscapedLessThanSignState},
		{'\u0000', scriptDataDoubleEscapedState, false, scriptDataDoubleEscapedState},
		{'a', scriptDataDoubleEscapedState, false, scriptDataDoubleEscapedState},
		{'$', scriptDataDoubleEscapedState, false, scriptDataDoubleEscapedState},

		{'-', scriptDataDoubleEscapedDashState, false, scriptDataDoubleEscapedDashDashState},
		{'<', scriptDataDoubleEscapedDashState, false, scriptDataDoubleEscapedLessThanSignState},
		{'\u0000', scriptDataDoubleEscapedDashState, false, scriptDataDoubleEscapedState},
		{'a', scriptDataDoubleEscapedDashState, false, scriptDataDoubleEscapedState},
		{'!', scriptDataDoubleEscapedDashState, false, scriptDataDoubleEscapedState},

		{'-', scriptDataDoubleEscapedDashDashState, false, scriptDataDoubleEscapedDashDashState},
		{'<', scriptDataDoubleEscapedDashDashState, false, scriptDataDoubleEscapedLessThanSignState},
		{'>', scriptDataDoubleEscapedDashDashState, false, scriptDataState},
		{'\u0000', scriptDataDoubleEscapedDashDashState, false, scriptDataDoubleEscapedState},
		{'a', scriptDataDoubleEscapedDashDashState, false, scriptDataDoubleEscapedState},
		{'!', scriptDataDoubleEscapedDashDashState, false, scriptDataDoubleEscapedState},

		{'/', scriptDataDoubleEscapedLessThanSignState, false, scriptDataDoubleEscapeEndState},
		{'a', scriptDataDoubleEscapedLessThanSignState, true, scriptDataDoubleEscapedState},
		{'#', scriptDataDoubleEscapedLessThanSignState, true, scriptDataDoubleEscapedState},

		//TODO: script data double escape end state has some side effects we need to write custom
		// tests for
		{'a', scriptDataDoubleEscapeEndState, false, scriptDataDoubleEscapeEndState},
		{'z', scriptDataDoubleEscapeEndState, false, scriptDataDoubleEscapeEndState},
		{'A', scriptDataDoubleEscapeEndState, false, scriptDataDoubleEscapeEndState},
		{'Z', scriptDataDoubleEscapeEndState, false, scriptDataDoubleEscapeEndState},
		{'@', scriptDataDoubleEscapeEndState, true, scriptDataDoubleEscapedState},
		{'1', scriptDataDoubleEscapeEndState, true, scriptDataDoubleEscapedState},

		{'\u0009', beforeAttributeNameState, false, beforeAttributeNameState},
		{'\u000A', beforeAttributeNameState, false, beforeAttributeNameState},
		{'\u000C', beforeAttributeNameState, false, beforeAttributeNameState},
		{'\u0020', beforeAttributeNameState, false, beforeAttributeNameState},
		{'/', beforeAttributeNameState, true, afterAttributeNameState},
		{'>', beforeAttributeNameState, true, afterAttributeNameState},
		{'=', beforeAttributeNameState, false, attributeNameState},
		{'!', beforeAttributeNameState, true, attributeNameState},
		{'@', beforeAttributeNameState, true, attributeNameState},
		{'a', beforeAttributeNameState, true, attributeNameState},

		{'\u0009', attributeNameState, true, afterAttributeNameState},
		{'\u000A', attributeNameState, true, afterAttributeNameState},
		{'\u000C', attributeNameState, true, afterAttributeNameState},
		{'\u0020', attributeNameState, true, afterAttributeNameState},
		{'\u002F', attributeNameState, true, afterAttributeNameState},
		{'\u003E', attributeNameState, true, afterAttributeNameState},
		{'=', attributeNameState, false, beforeAttributeValueState},
		{'A', attributeNameState, false, attributeNameState},
		{'B', attributeNameState, false, attributeNameState},
		{'\u0000', attributeNameState, false, attributeNameState},
		{'"', attributeNameState, false, attributeNameState},
		{'\'', attributeNameState, false, attributeNameState},
		{'<', attributeNameState, false, attributeNameState},
		{'a', attributeNameState, false, attributeNameState},
		{'1', attributeNameState, false, attributeNameState},
		{'@', attributeNameState, false, attributeNameState},

		{'\u0009', afterAttributeNameState, false, afterAttributeNameState},
		{'\u000A', afterAttributeNameState, false, afterAttributeNameState},
		{'\u000C', afterAttributeNameState, false, afterAttributeNameState},
		{'\u0020', afterAttributeNameState, false, afterAttributeNameState},
		{'/', afterAttributeNameState, false, selfClosingStartTagState},
		{'=', afterAttributeNameState, false, beforeAttributeValueState},
		{'>', afterAttributeNameState, false, dataState},
		{'a', afterAttributeNameState, true, attributeNameState},
		{'B', afterAttributeNameState, true, attributeNameState},
		{'C', afterAttributeNameState, true, attributeNameState},
		{'1', afterAttributeNameState, true, attributeNameState},
		{'%', afterAttributeNameState, true, attributeNameState},

		{'\u0009', beforeAttributeValueState, false, beforeAttributeValueState},
		{'\u000A', beforeAttributeValueState, false, beforeAttributeValueState},
		{'\u000C', beforeAttributeValueState, false, beforeAttributeValueState},
		{'\u0020', beforeAttributeValueState, false, beforeAttributeValueState},
		{'"', beforeAttributeValueState, false, attributeValueDoubleQuotedState},
		{'\'', beforeAttributeValueState, false, attributeValueSingleQuotedState},
		{'>', beforeAttributeValueState, false, dataState},
		{'a', beforeAttributeValueState, true, attributeValueUnquotedState},
		{'1', beforeAttributeValueState, true, attributeValueUnquotedState},

		{'"', attributeValueDoubleQuotedState, false, afterAttributeValueQuotedState},
		{'&', attributeValueDoubleQuotedState, false, characterReferenceState},
		{'\u0000', attributeValueDoubleQuotedState, false, attributeValueDoubleQuotedState},
		{'a', attributeValueDoubleQuotedState, false, attributeValueDoubleQuotedState},
		{'$', attributeValueDoubleQuotedState, false, attributeValueDoubleQuotedState},

		{'\'', attributeValueSingleQuotedState, false, afterAttributeValueQuotedState},
		{'&', attributeValueSingleQuotedState, false, characterReferenceState},
		{'\u0000', attributeValueSingleQuotedState, false, attributeValueSingleQuotedState},
		{'a', attributeValueSingleQuotedState, false, attributeValueSingleQuotedState},
		{'5', attributeValueSingleQuotedState, false, attributeValueSingleQuotedState},

		{'\u0009', attributeValueUnquotedState, false, beforeAttributeNameState},
		{'\u000A', attributeValueUnquotedState, false, beforeAttributeNameState},
		{'\u000C', attributeValueUnquotedState, false, beforeAttributeNameState},
		{'\u0020', attributeValueUnquotedState, false, beforeAttributeNameState},
		{'&', attributeValueUnquotedState, false, characterReferenceState},
		{'>', attributeValueUnquotedState, false, dataState},
		{'\u0000', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'"', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'\'', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'<', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'=', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'`', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'A', attributeValueUnquotedState, false, attributeValueUnquotedState},
		{'%', attributeValueUnquotedState, false, attributeValueUnquotedState},

		{'\u0009', afterAttributeValueQuotedState, false, beforeAttributeNameState},
		{'\u000A', afterAttributeValueQuotedState, false, beforeAttributeNameState},
		{'\u000C', afterAttributeValueQuotedState, false, beforeAttributeNameState},
		{'\u0020', afterAttributeValueQuotedState, false, beforeAttributeNameState},
		{'/', afterAttributeValueQuotedState, false, selfClosingStartTagState},
		{'>', afterAttributeValueQuotedState, false, dataState},
		{'A', afterAttributeValueQuotedState, true, beforeAttributeNameState},
		{'(', afterAttributeValueQuotedState, true, beforeAttributeNameState},

		{'>', selfClosingStartTagState, false, dataState},
		{'a', selfClosingStartTagState, true, beforeAttributeNameState},
		{'b', selfClosingStartTagState, true, beforeAttributeNameState},
		{'1', selfClosingStartTagState, true, beforeAttributeNameState},
		{'$', selfClosingStartTagState, true, beforeAttributeNameState},
		{'"', selfClosingStartTagState, true, beforeAttributeNameState},

		{'>', bogusCommentState, false, dataState},
		{'\u0000', bogusCommentState, false, bogusCommentState},
		{'a', bogusCommentState, false, bogusCommentState},
		{'A', bogusCommentState, false, bogusCommentState},
		{'1', bogusCommentState, false, bogusCommentState},
		{'#', bogusCommentState, false, bogusCommentState},
		{'^', bogusCommentState, false, bogusCommentState},

		// TODO: markup tests are going to require multiple characters. not sure how to test that like this

		{'-', commentStartState, false, commentStartDashState},
		{'>', commentStartState, false, dataState},
		{'A', commentStartState, true, commentState},
		{'(', commentStartState, true, commentState},

		{'-', commentStartDashState, false, commentEndState},
		{'>', commentStartDashState, false, dataState},
		{'A', commentStartDashState, true, commentState},
		{'(', commentStartDashState, true, commentState},

		{'<', commentState, false, commentLessThanSignState},
		{'-', commentState, false, commentEndDashState},
		{'\u0000', commentState, false, commentState},
		{'A', commentState, false, commentState},
		{')', commentState, false, commentState},

		{'!', commentLessThanSignState, false, commentLessThanSignBangState},
		{'<', commentLessThanSignState, false, commentLessThanSignState},
		{'A', commentLessThanSignState, true, commentState},
		{'*', commentLessThanSignState, true, commentState},

		{'-', commentLessThanSignBangState, false, commentLessThanSignBangDashState},
		{'A', commentLessThanSignBangState, true, commentState},
		{'@', commentLessThanSignBangState, true, commentState},

		{'-', commentLessThanSignBangDashState, false, commentLessThanSignBangDashDashState},
		{'A', commentLessThanSignBangDashState, true, commentEndDashState},
		{'!', commentLessThanSignBangDashState, true, commentEndDashState},

		{'>', commentLessThanSignBangDashDashState, true, commentEndState},
		{'A', commentLessThanSignBangDashDashState, true, commentEndState},
		{'^', commentLessThanSignBangDashDashState, true, commentEndState},

		{'-', commentEndDashState, false, commentEndState},
		{'A', commentEndDashState, true, commentState},
		{'#', commentEndDashState, true, commentState},

		{'>', commentEndState, false, dataState},
		{'!', commentEndState, false, commentEndBangState},
		{'-', commentEndState, false, commentEndState},
		{'A', commentEndState, true, commentState},
		{'(', commentEndState, true, commentState},

		{'-', commentEndBangState, false, commentEndDashState},
		{'>', commentEndBangState, false, dataState},
		{'A', commentEndBangState, true, commentState},
		{'*', commentEndBangState, true, commentState},

		{'\u0009', doctypeState, false, beforeDoctypeNameState},
		{'\u000A', doctypeState, false, beforeDoctypeNameState},
		{'\u000C', doctypeState, false, beforeDoctypeNameState},
		{'\u0020', doctypeState, false, beforeDoctypeNameState},
		{'>', doctypeState, true, beforeDoctypeNameState},
		{'a', doctypeState, true, beforeDoctypeNameState},
		{'&', doctypeState, true, beforeDoctypeNameState},

		{'\u0009', beforeDoctypeNameState, false, beforeDoctypeNameState},
		{'\u000A', beforeDoctypeNameState, false, beforeDoctypeNameState},
		{'\u000C', beforeDoctypeNameState, false, beforeDoctypeNameState},
		{'\u0020', beforeDoctypeNameState, false, beforeDoctypeNameState},
		{'A', beforeDoctypeNameState, false, doctypeNameState},
		{'Z', beforeDoctypeNameState, false, doctypeNameState},
		{'\u0000', beforeDoctypeNameState, false, doctypeNameState},
		{'>', beforeDoctypeNameState, false, dataState},
		{'a', beforeDoctypeNameState, false, doctypeNameState},
		{'*', beforeDoctypeNameState, false, doctypeNameState},

		{'\u0009', doctypeNameState, false, afterDoctypeNameState},
		{'\u000A', doctypeNameState, false, afterDoctypeNameState},
		{'\u000C', doctypeNameState, false, afterDoctypeNameState},
		{'\u0020', doctypeNameState, false, afterDoctypeNameState},
		{'A', doctypeNameState, false, doctypeNameState},
		{'Z', doctypeNameState, false, doctypeNameState},
		{'\u0000', doctypeNameState, false, doctypeNameState},
		{'a', doctypeNameState, false, doctypeNameState},
		{'*', doctypeNameState, false, doctypeNameState},

		//TODO: after doctype name state has some weird stateful actions at the end that will be hard to
		// test this way

		{'\u0009', afterDoctypeNameState, false, afterDoctypeNameState},
		{'\u000A', afterDoctypeNameState, false, afterDoctypeNameState},
		{'\u000C', afterDoctypeNameState, false, afterDoctypeNameState},
		{'\u0020', afterDoctypeNameState, false, afterDoctypeNameState},
		{'>', afterDoctypeNameState, false, dataState},

		{'\u0009', afterDoctypePublicKeywordState, false, beforeDoctypePublicIdentifierState},
		{'\u000A', afterDoctypePublicKeywordState, false, beforeDoctypePublicIdentifierState},
		{'\u000C', afterDoctypePublicKeywordState, false, beforeDoctypePublicIdentifierState},
		{'\u0020', afterDoctypePublicKeywordState, false, beforeDoctypePublicIdentifierState},
		{'"', afterDoctypePublicKeywordState, false, doctypePublicIdentifierDoubleQuotedState},
		{'\'', afterDoctypePublicKeywordState, false, doctypePublicIdentifierSingleQuotedState},
		{'>', afterDoctypePublicKeywordState, false, dataState},
		{'a', afterDoctypePublicKeywordState, true, bogusDoctypeState},
		{'&', afterDoctypePublicKeywordState, true, bogusDoctypeState},

		{'\u0009', beforeDoctypePublicIdentifierState, false, beforeDoctypePublicIdentifierState},
		{'\u000A', beforeDoctypePublicIdentifierState, false, beforeDoctypePublicIdentifierState},
		{'\u000C', beforeDoctypePublicIdentifierState, false, beforeDoctypePublicIdentifierState},
		{'\u0020', beforeDoctypePublicIdentifierState, false, beforeDoctypePublicIdentifierState},
		{'"', beforeDoctypePublicIdentifierState, false, doctypePublicIdentifierDoubleQuotedState},
		{'\'', beforeDoctypePublicIdentifierState, false, doctypePublicIdentifierSingleQuotedState},
		{'>', beforeDoctypePublicIdentifierState, false, dataState},
		{'a', beforeDoctypePublicIdentifierState, true, bogusDoctypeState},
		{'(', beforeDoctypePublicIdentifierState, true, bogusDoctypeState},

		{'"', doctypePublicIdentifierDoubleQuotedState, false, afterDoctypePublicIdentifierState},
		{'\u0000', doctypePublicIdentifierDoubleQuotedState, false, doctypePublicIdentifierDoubleQuotedState},
		{'>', doctypePublicIdentifierDoubleQuotedState, false, dataState},
		{'a', doctypePublicIdentifierDoubleQuotedState, false, doctypePublicIdentifierDoubleQuotedState},
		{'*', doctypePublicIdentifierDoubleQuotedState, false, doctypePublicIdentifierDoubleQuotedState},

		{'\'', doctypePublicIdentifierSingleQuotedState, false, afterDoctypePublicIdentifierState},
		{'\u0000', doctypePublicIdentifierSingleQuotedState, false, doctypePublicIdentifierSingleQuotedState},
		{'>', doctypePublicIdentifierSingleQuotedState, false, dataState},
		{'a', doctypePublicIdentifierSingleQuotedState, false, doctypePublicIdentifierSingleQuotedState},
		{'(', doctypePublicIdentifierSingleQuotedState, false, doctypePublicIdentifierSingleQuotedState},

		{'\u0009', afterDoctypePublicIdentifierState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'\u000A', afterDoctypePublicIdentifierState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'\u000C', afterDoctypePublicIdentifierState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'\u0020', afterDoctypePublicIdentifierState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'>', afterDoctypePublicIdentifierState, false, dataState},
		{'"', afterDoctypePublicIdentifierState, false, doctypeSystemIdentifierDoubleQuotedState},
		{'\'', afterDoctypePublicIdentifierState, false, doctypeSystemIdentifierSingleQuotedState},
		{'A', afterDoctypePublicIdentifierState, true, bogusDoctypeState},
		{'(', afterDoctypePublicIdentifierState, true, bogusDoctypeState},

		{'\u0009', betweenDoctypePublicAndSystemIdentifiersState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'\u000A', betweenDoctypePublicAndSystemIdentifiersState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'\u000C', betweenDoctypePublicAndSystemIdentifiersState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'\u0020', betweenDoctypePublicAndSystemIdentifiersState, false, betweenDoctypePublicAndSystemIdentifiersState},
		{'>', betweenDoctypePublicAndSystemIdentifiersState, false, dataState},
		{'"', betweenDoctypePublicAndSystemIdentifiersState, false, doctypeSystemIdentifierDoubleQuotedState},
		{'\'', betweenDoctypePublicAndSystemIdentifiersState, false, doctypeSystemIdentifierSingleQuotedState},
		{'A', betweenDoctypePublicAndSystemIdentifiersState, true, bogusDoctypeState},
		{'#', betweenDoctypePublicAndSystemIdentifiersState, true, bogusDoctypeState},

		{'\u0009', afterDoctypeSystemKeywordState, false, beforeDoctypeSystemIdentifierState},
		{'\u000A', afterDoctypeSystemKeywordState, false, beforeDoctypeSystemIdentifierState},
		{'\u000C', afterDoctypeSystemKeywordState, false, beforeDoctypeSystemIdentifierState},
		{'\u0020', afterDoctypeSystemKeywordState, false, beforeDoctypeSystemIdentifierState},
		{'"', afterDoctypeSystemKeywordState, false, doctypeSystemIdentifierDoubleQuotedState},
		{'\'', afterDoctypeSystemKeywordState, false, doctypeSystemIdentifierSingleQuotedState},
		{'>', afterDoctypeSystemKeywordState, false, dataState},
		{'A', afterDoctypeSystemKeywordState, true, bogusDoctypeState},
		{'$', afterDoctypeSystemKeywordState, true, bogusDoctypeState},

		{'\u0009', beforeDoctypeSystemIdentifierState, false, beforeDoctypeSystemIdentifierState},
		{'\u000A', beforeDoctypeSystemIdentifierState, false, beforeDoctypeSystemIdentifierState},
		{'\u000C', beforeDoctypeSystemIdentifierState, false, beforeDoctypeSystemIdentifierState},
		{'\u0020', beforeDoctypeSystemIdentifierState, false, beforeDoctypeSystemIdentifierState},
		{'"', beforeDoctypeSystemIdentifierState, false, doctypeSystemIdentifierDoubleQuotedState},
		{'\'', beforeDoctypeSystemIdentifierState, false, doctypeSystemIdentifierSingleQuotedState},
		{'>', beforeDoctypeSystemIdentifierState, false, dataState},
		{'A', beforeDoctypeSystemIdentifierState, true, bogusDoctypeState},
		{'*', beforeDoctypeSystemIdentifierState, true, bogusDoctypeState},

		{'"', doctypeSystemIdentifierDoubleQuotedState, false, afterDoctypeSystemIdentifierState},
		{'\u0000', doctypeSystemIdentifierDoubleQuotedState, false, doctypeSystemIdentifierDoubleQuotedState},
		{'>', doctypeSystemIdentifierDoubleQuotedState, false, dataState},
		{'A', doctypeSystemIdentifierDoubleQuotedState, false, doctypeSystemIdentifierDoubleQuotedState},
		{'@', doctypeSystemIdentifierDoubleQuotedState, false, doctypeSystemIdentifierDoubleQuotedState},

		{'\'', doctypeSystemIdentifierSingleQuotedState, false, afterDoctypeSystemIdentifierState},
		{'\u0000', doctypeSystemIdentifierSingleQuotedState, false, doctypeSystemIdentifierSingleQuotedState},
		{'>', doctypeSystemIdentifierSingleQuotedState, false, dataState},
		{'A', doctypeSystemIdentifierSingleQuotedState, false, doctypeSystemIdentifierSingleQuotedState},
		{'*', doctypeSystemIdentifierSingleQuotedState, false, doctypeSystemIdentifierSingleQuotedState},

		{'\u0009', afterDoctypeSystemIdentifierState, false, afterDoctypeSystemIdentifierState},
		{'\u000A', afterDoctypeSystemIdentifierState, false, afterDoctypeSystemIdentifierState},
		{'\u000C', afterDoctypeSystemIdentifierState, false, afterDoctypeSystemIdentifierState},
		{'\u0020', afterDoctypeSystemIdentifierState, false, afterDoctypeSystemIdentifierState},
		{'>', afterDoctypeSystemIdentifierState, false, dataState},
		{'A', afterDoctypeSystemIdentifierState, true, bogusDoctypeState},
		{'#', afterDoctypeSystemIdentifierState, true, bogusDoctypeState},

		{'>', bogusDoctypeState, false, dataState},
		{'\u0000', bogusDoctypeState, false, bogusDoctypeState},
		{'a', bogusDoctypeState, false, bogusDoctypeState},
		{'(', bogusDoctypeState, false, bogusDoctypeState},

		{']', cdataSectionState, false, cdataSectionBracketState},
		{'A', cdataSectionState, false, cdataSectionState},
		{'^', cdataSectionState, false, cdataSectionState},

		{']', cdataSectionBracketState, false, cdataSectionEndState},
		{'A', cdataSectionBracketState, true, cdataSectionState},
		{'[', cdataSectionBracketState, true, cdataSectionState},

		{']', cdataSectionEndState, false, cdataSectionEndState},
		{'>', cdataSectionEndState, false, dataState},
		{'A', cdataSectionEndState, true, cdataSectionState},
		{'[', cdataSectionEndState, true, cdataSectionState},

		{'a', characterReferenceState, true, namedCharacterReferenceState},
		{'b', characterReferenceState, true, namedCharacterReferenceState},
		{'1', characterReferenceState, true, namedCharacterReferenceState},
		{'2', characterReferenceState, true, namedCharacterReferenceState},
		{'#', characterReferenceState, false, numericCharacterReferenceState},
		{'/', characterReferenceState, true, dataState}, // the default return state in the tests
		{'"', characterReferenceState, true, dataState}, // the default return state in the tests

		//TODO: named character reference state needs specific tests because it consumes a number of characters

		{'a', ambiguousAmpersandState, false, ambiguousAmpersandState},
		{'3', ambiguousAmpersandState, false, ambiguousAmpersandState},

		//TODO: requires we set up a return state
		/*{';', ambiguousAmpersandState, false, p.returnState},
		{'[', ambiguousAmpersandState, true, p.returnState},
		{'@', ambiguousAmpersandState, true, p.returnState},*/

		{'\u0078', numericCharacterReferenceState, false, hexadecimalCharacterReferenceStartState},
		{'\u0058', numericCharacterReferenceState, false, hexadecimalCharacterReferenceStartState},
		{'[', numericCharacterReferenceState, true, decimalCharacterReferenceStartState},
		{'a', numericCharacterReferenceState, true, decimalCharacterReferenceStartState},

		{'a', hexadecimalCharacterReferenceStartState, true, hexadecimalCharacterReferenceState},
		{'b', hexadecimalCharacterReferenceStartState, true, hexadecimalCharacterReferenceState},
		// TODO: requires a return state
		/*{'p', hexadecimalCharacterReferenceStartState, true, p.returnState},*/

		{'0', decimalCharacterReferenceStartState, true, decimalCharacterReferenceState},
		{'1', decimalCharacterReferenceStartState, true, decimalCharacterReferenceState},
		// TODO: requires a return state
		/*{'A', decimalCharacterReferenceStartState, true, p.returnState},*/

		{'0', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'1', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'9', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'A', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'F', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'B', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'a', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'b', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{'F', hexadecimalCharacterReferenceState, false, hexadecimalCharacterReferenceState},
		{';', hexadecimalCharacterReferenceState, false, numericCharacterReferenceEndState},
		{'#', hexadecimalCharacterReferenceState, true, numericCharacterReferenceEndState},
		{']', hexadecimalCharacterReferenceState, true, numericCharacterReferenceEndState},

		{'0', decimalCharacterReferenceState, false, decimalCharacterReferenceState},
		{'1', decimalCharacterReferenceState, false, decimalCharacterReferenceState},
		{'9', decimalCharacterReferenceState, false, decimalCharacterReferenceState},
		{';', decimalCharacterReferenceState, false, numericCharacterReferenceEndState},
		{'g', decimalCharacterReferenceState, true, numericCharacterReferenceEndState},
		{'!', decimalCharacterReferenceState, true, numericCharacterReferenceEndState},
		{'a', decimalCharacterReferenceState, true, numericCharacterReferenceEndState},

		// TODO: numeric character reference end state will need separate tests
	}

	for _, tt := range stateParserTests {
		runStateParserTest(tt, t)
	}
}

// helper function to parallelize the above test case
func runStateParserTest(testcase stateMachineTestCase, t *testing.T) {
	testName := fmt.Sprintf("%s-%#U", testcase.startingState, testcase.inRune)
	t.Run(testName, func(t *testing.T) {
		t.Parallel()
		p := NewHTMLTokenizer(strings.NewReader(""))
		reconsume, state := p.stateToParser(testcase.startingState)(testcase.inRune, false)
		if state != testcase.nextExpectedState {
			t.Errorf("Expected %d state, got %d", testcase.nextExpectedState, state)
		}

		if reconsume != testcase.shouldReconsume {
			t.Errorf("Expected to reconsume to be %+v", reconsume)
		}
	})
}

type parserStatefulnessTestCase struct {
	inHTML     string                                // the HTML to tokenize
	startState tokenizerState                        // the starting state of the tokenizer
	testFunc   func(*HTMLTokenizer) (string, string) // since we are testing internal state, we need a function that can look inside the tokenizer
	setup      func(*HTMLTokenizer)                  // any setup code will be run before tokenization
}

// TestParseStatefulness is a set of tests that each run the full tokenizer minus any EOF handlers.
// These tests are good for checking the internal state of the token builder at various points. We don't
// use the EOF handlers because the EOF handlers will often erase the token builder state.
func TestParseStatefulness(t *testing.T) {
	parserStatefulnessTestCases := []parserStatefulnessTestCase{
		{"&", dataState, func(p *HTMLTokenizer) (string, string) { return p.returnState.String(), dataState.String() }, nil},
		{"&", rcDataState, func(p *HTMLTokenizer) (string, string) { return p.returnState.String(), rcDataState.String() }, nil},
		{"b", tagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "b" }, nil},
		{"ba", tagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "ba" }, nil},
		{"bAc", tagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "bac" }, nil},
		{"bA\u0000c", tagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "ba\uFFFDc" }, nil},
		{"a", endTagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "a" }, nil},
		{"P", endTagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "p" }, nil},
		{"1", endTagOpenState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "1" }, nil},
		{"U", tagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"u", tagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"\u0000", tagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "\uFFFD" }, nil},
		{"<", tagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "<" }, nil},
		{"U", rcDataEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"u", rcDataEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"U", rawTextEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"u", rawTextEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"U", scriptDataEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"u", scriptDataEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"U", scriptDataEscapedEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"u", scriptDataEscapedEndTagNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "u" }, nil},
		{"U", scriptDataDoubleEscapeStartState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.TempBuffer(), "u" }, nil},
		{"u", scriptDataDoubleEscapeStartState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.TempBuffer(), "u" }, nil},
		{"U", scriptDataDoubleEscapeEndState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.TempBuffer(), "u" }, nil},
		{"u", scriptDataDoubleEscapeEndState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.TempBuffer(), "u" }, nil},
		{"U", attributeNameState, func(p *HTMLTokenizer) (string, string) {
			_, ok := p.tokenBuilder.attributes["u"]
			return fmt.Sprintf("%t", ok), "true"
		}, nil},
		{"u", attributeNameState, func(p *HTMLTokenizer) (string, string) {
			_, ok := p.tokenBuilder.attributes["u"]
			return fmt.Sprintf("%t", ok), "true"
		}, nil},
		{"1", attributeNameState, func(p *HTMLTokenizer) (string, string) {
			_, ok := p.tokenBuilder.attributes["1"]
			return fmt.Sprintf("%t", ok), "true"
		}, nil},
		{"\u0000", attributeNameState, func(p *HTMLTokenizer) (string, string) {
			_, ok := p.tokenBuilder.attributes["\uFFFD"]
			return fmt.Sprintf("%t", ok), "true"
		}, nil},
		{"\u0000", attributeValueDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "\uFFFD" }, nil},
		{"a", attributeValueDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "a" }, nil},
		{"A", attributeValueDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "A" }, nil},
		{"\u0000", attributeValueSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "\uFFFD" }, nil},
		{"a", attributeValueSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "a" }, nil},
		{"A", attributeValueSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "A" }, nil},
		{"\u0000", attributeValueUnquotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "\uFFFD" }, nil},
		{"a", attributeValueUnquotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "a" }, nil},
		{"A", attributeValueUnquotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "A" }, nil},
		{"&", attributeValueUnquotedState, func(p *HTMLTokenizer) (string, string) {
			return p.returnState.String(), attributeValueUnquotedState.String()
		}, nil},
		{">", selfClosingStartTagState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.selfClosing), "true" }, nil},
		{"\u0000", bogusCommentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "\uFFFD" }, nil},
		{"a", bogusCommentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "a" }, nil},
		{"A", bogusCommentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "A" }, nil},
		{"!", bogusCommentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "!" }, nil},
		{"3", bogusCommentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "3" }, nil},
		{"3", commentStartDashState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-3" }, nil},
		{"a", commentStartDashState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-a" }, nil},
		{"#", commentStartDashState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-#" }, nil},
		{"<", commentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "<" }, nil},
		{"\u0000", commentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "\uFFFD" }, nil},
		{"!", commentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "!" }, nil},
		{"a", commentState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "a" }, nil},
		{"!", commentLessThanSignState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "!" }, nil},
		{"<", commentLessThanSignState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "<" }, nil},
		{"a", commentEndDashState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-a" }, nil},
		{"1", commentEndDashState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-1" }, nil},
		{"!", commentEndDashState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-!" }, nil},
		{"-", commentEndState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "-" }, nil},
		{"A", commentEndState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "--A" }, nil},
		{"@", commentEndState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "--@" }, nil},
		{"-", commentEndBangState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "--!" }, nil},
		{"@", commentEndBangState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.data.String(), "--!@" }, nil},
		{"A", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "a" }, nil},
		{"a", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "a" }, nil},
		{"\u0000", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "\uFFFD" }, nil},
		{">", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"1", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "1" }, nil},
		{"A", doctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "a" }, nil},
		{"a", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "a" }, nil},
		{"\u0000", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "\uFFFD" }, nil},
		{"!", beforeDoctypeNameState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.name.String(), "!" }, nil},
		{">", beforeDoctypePublicIdentifierState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"A", beforeDoctypePublicIdentifierState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"\u0000", doctypePublicIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.publicID.String(), "\uFFFD" }, nil},
		{">", doctypePublicIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"A", doctypePublicIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.publicID.String(), "A" }, nil},
		{"\u0000", doctypePublicIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.publicID.String(), "\uFFFD" }, nil},
		{">", doctypePublicIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"A", doctypePublicIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.publicID.String(), "A" }, nil},
		{"A", afterDoctypePublicIdentifierState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"!", afterDoctypePublicIdentifierState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"!", betweenDoctypePublicAndSystemIdentifiersState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{">", afterDoctypeSystemKeywordState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"!", afterDoctypeSystemKeywordState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{">", beforeDoctypeSystemIdentifierState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"!", beforeDoctypeSystemIdentifierState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"\u0000", doctypeSystemIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.systemID.String(), "\uFFFD" }, nil},
		{"a", doctypeSystemIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.systemID.String(), "a" }, nil},
		{"!", doctypeSystemIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.systemID.String(), "!" }, nil},
		{">", doctypeSystemIdentifierDoubleQuotedState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},
		{"\u0000", doctypeSystemIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.systemID.String(), "\uFFFD" }, nil},
		{"a", doctypeSystemIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.systemID.String(), "a" }, nil},
		{"!", doctypeSystemIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.systemID.String(), "!" }, nil},
		{">", doctypeSystemIdentifierSingleQuotedState, func(p *HTMLTokenizer) (string, string) { return fmt.Sprintf("%t", p.tokenBuilder.forceQuirks), "true" }, nil},

		//character reference come back
		// same with named character reference
		{"a", ambiguousAmpersandState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "a" },
			func(p *HTMLTokenizer) {
				p.returnState = attributeValueDoubleQuotedState
			}},
		{"A", ambiguousAmpersandState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "A" }, func(p *HTMLTokenizer) {
			p.returnState = attributeValueDoubleQuotedState
		}},
		{"1", ambiguousAmpersandState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.attributeValue.String(), "1" }, func(p *HTMLTokenizer) {
			p.returnState = attributeValueDoubleQuotedState
		}},
		{"\u0058", numericCharacterReferenceState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.TempBuffer(), "\u0058" }, nil},
		{"\u0078", numericCharacterReferenceState, func(p *HTMLTokenizer) (string, string) { return p.tokenBuilder.TempBuffer(), "\u0078" }, nil},
		{"1", hexadecimalCharacterReferenceState, func(p *HTMLTokenizer) (string, string) {
			return fmt.Sprintf("%d", p.tokenBuilder.GetCharRef()), "1"
		}, nil},
		{"22", hexadecimalCharacterReferenceState, func(p *HTMLTokenizer) (string, string) {
			return fmt.Sprintf("%d", p.tokenBuilder.GetCharRef()), "34"
		}, nil},
		{"FF", hexadecimalCharacterReferenceState, func(p *HTMLTokenizer) (string, string) {
			return fmt.Sprintf("%d", p.tokenBuilder.GetCharRef()), "255"
		}, nil},
		{"ff", hexadecimalCharacterReferenceState, func(p *HTMLTokenizer) (string, string) {
			return fmt.Sprintf("%d", p.tokenBuilder.GetCharRef()), "255"
		}, nil},
		{"134", decimalCharacterReferenceState, func(p *HTMLTokenizer) (string, string) {
			return fmt.Sprintf("%d", p.tokenBuilder.GetCharRef()), "8224"
		}, nil},

		// come back to numeric character reference end state
	}

	for _, testcase := range parserStatefulnessTestCases {
		runParserStatefulnessTest(testcase, t)
	}
}

// helper function to paralleize the above tests
func runParserStatefulnessTest(testcase parserStatefulnessTestCase, t *testing.T) {
	testName := fmt.Sprintf("%s-%s", testcase.startState, testcase.inHTML)
	t.Run(testName, func(t *testing.T) {
		t.Parallel()
		p := NewParser(strings.NewReader(testcase.inHTML))
		// run through all the runes and then check the state at the end.
		// I'm also wanting to start a certain state so that is what I am doing with this function
		// helper.
		if testcase.setup != nil {
			testcase.setup(p.Tokenizer)
		}
		p.startAt(&testcase.startState)
		answer, expected := testcase.testFunc(p.Tokenizer)
		if expected != answer {
			t.Errorf("Expected %s, but got %s", expected, answer)
		}
	})
}

//TODO: implement these tests: https://github.com/html5lib/html5lib-test
type HTML5Tests struct {
	Tests []HTML5Test `json:"tests"`
}

type HTML5Test struct {
	Description   string          `json:"description"`
	Input         string          `json:"input"`
	Output        [][]interface{} `json:"output"`
	DoubleEscaped bool            `json:doubleEscaped`
	LastStartTag  string          `json:lastStartTag`
	Errors        []struct {
		Code string `json:"code"`
		Line int    `json:"line"`
		Col  int    `json:"col"`
	} `json:"errors,omitempty"`
	InitialStates []string `json:"initialStates,omitempty"`
}

func TestHTML5Lib(t *testing.T) {
	allTests := &HTML5Tests{
		Tests: make([]HTML5Test, 0),
	}
	dir := "./tests/tokenizer/"
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		t.Error(err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".test") {
			continue
		}
		path := filepath.Join(dir, file.Name())
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Error(err)
		}

		var tests *HTML5Tests
		err = json.Unmarshal(data, &tests)
		if err != nil {
			t.Error(err)
			return
		}

		allTests.Tests = append(allTests.Tests, tests.Tests...)
	}

	for _, test := range allTests.Tests {
		runHTML5Test(test, t)
	}
}

func getInitState(state string) (tokenizerState, error) {
	switch state {
	case "Data state":
		return dataState, nil
	case "PLAINTEXT state":
		return plaintextState, nil
	case "RCDATA state":
		return rcDataState, nil
	case "RAWTEXT state":
		return rawTextState, nil
	case "Script data state":
		return scriptDataState, nil
	case "CDATA section state":
		return cdataSectionState, nil
	default:
		return dataState, fmt.Errorf("invalid tokenizer state %s", state)
	}
}

func formatString(v interface{}, de bool) string {
	if de {
		s, err := doubleEscape(v.(string))
		if err != nil {
			return ""
		}
		return s
	}

	return v.(string)
}

func formatOutputs(outputs [][]interface{}, doubleEscape bool) []Token {
	tb := newTokenBuilder()
	tokens := []Token{}

	publicID := missing
	systemID := missing
	var name string
	for _, v := range outputs {
		tb.NewToken()
		if len(v) == 0 {
			continue
		}
		switch v[0].(string) {
		case "DOCTYPE":
			if len(v) >= 2 {
				if v[1] != nil {
					name = formatString(v[1], doubleEscape)
				}
			}

			if len(v) >= 3 && v[2] != nil {
				publicID = formatString(v[2], doubleEscape)
			}

			if len(v) >= 4 && v[3] != nil {
				systemID = formatString(v[3], doubleEscape)
			}

			correctness := v[4].(bool)
			if !correctness {
				tb.EnableForceQuirks()
			}

			for _, v := range name {
				tb.WriteName(v)
			}
			if publicID == "" {
				tb.WritePublicIdentifierEmpty()
			} else if publicID != missing {
				for _, v := range publicID {
					tb.WritePublicIdentifier(v)
				}
			}

			if systemID == "" {
				tb.WriteSystemIdentifierEmpty()
			} else if systemID != missing {
				for _, v := range systemID {
					tb.WriteSystemIdentifier(v)
				}
			}
			tokens = append(tokens, *tb.DocTypeToken())
		case "StartTag":
			if len(v) >= 2 && v[1] != nil {
				for _, n := range v[1].(string) {
					tb.WriteName(n)
				}

			}

			if len(v) >= 3 && v[2] != nil {
				for name, value := range v[2].(map[string]interface{}) {
					for _, n := range name {
						tb.WriteAttributeName(n)
					}
					for _, r := range value.(string) {
						tb.WriteAttributeValue(r)
					}
					tb.CommitAttribute()
				}
			}

			if len(v) >= 4 && v[3] != nil {
				if v[3].(bool) {
					tb.EnableSelfClosing()
				}
			}

			tokens = append(tokens, *tb.StartTagToken())
		case "EndTag":
			if len(v) >= 2 && v[1] != nil {
				s := formatString(v[1], doubleEscape)
				for _, r := range s {
					tb.WriteName(r)
				}
			}

			tokens = append(tokens, *tb.EndTagToken())
		case "Comment":
			if len(v) >= 1 && v[1] != nil {
				s := formatString(v[1], doubleEscape)
				for _, r := range s {
					tb.WriteData(r)
				}
			}
			tokens = append(tokens, *tb.CommentToken())
		case "Character":
			if len(v) >= 2 && v[1] != nil {
				s := formatString(v[1], doubleEscape)
				for _, r := range s {
					tokens = append(tokens, *tb.CharacterToken(r))
				}
			}
		}
	}
	return tokens
}

func doubleEscape(s string) (string, error) {
	ns := strconv.QuoteToASCII(s)
	rs := strings.ReplaceAll(ns, "\\\\", "\\")

	n, err := strconv.Unquote(rs)
	if err != nil {
		return s, err
	}

	return n, nil
}

func runHTML5Test(test HTML5Test, t *testing.T) {
	t.Run(test.Description, func(t *testing.T) {
		t.Parallel()
		if test.DoubleEscaped {
			var err error
			test.Input, err = doubleEscape(test.Input)
			if err != nil {
				t.Error(err)
				return
			}
		}

		if len(test.InitialStates) == 0 {
			test.InitialStates = []string{"Data state"}
		}
		expectedTokens := formatOutputs(test.Output, test.DoubleEscaped)
		for _, initState := range test.InitialStates {
			p := NewParser(strings.NewReader(test.Input))

			iState, err := getInitState(initState)
			if err != nil {
				t.Fatal(err)
			}

			if test.LastStartTag != "" {
				p.Tokenizer.lastEmittedStartTagName = test.LastStartTag
			}

			tokens, err := p.startAt(&iState)
			if err != nil {
				t.Fatal(err)
			}
			// the expected tokens don't include the EOF token, but `tokens` does
			tokens = tokens[:len(tokens)-1]
			if len(tokens) != len(expectedTokens) {
				t.Fatalf("Unexpected number of tokens. Expected %d, got %d", len(expectedTokens), len(tokens))
			}
			for i, token := range tokens {
				if !token.Equal(&expectedTokens[i]) {
					t.Fatalf("Got the wrong token. Expected %s, got %s", &expectedTokens[i], token)
				}
			}
		}
	})
}
