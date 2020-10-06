// Code generated by "stringer -type=parseError"; DO NOT EDIT.

package parser

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[noError-0]
	_ = x[abruptClosingOfEmptyComment-1]
	_ = x[abruptDoctypePublicIdentifier-2]
	_ = x[abruptDoctypeSystemIdentifier-3]
	_ = x[absenceOfDigitsInNumericCharacterReference-4]
	_ = x[cdataInHTMLContent-5]
	_ = x[characterReferenceOutsideUnicodeRange-6]
	_ = x[controlCharacterInInputSteam-7]
	_ = x[controlCharacterReference-8]
	_ = x[endTagWithAttributes-9]
	_ = x[duplicateAttribute-10]
	_ = x[endTagWithTrailingSolidus-11]
	_ = x[eofBeforeTagName-12]
	_ = x[eofInCdata-13]
	_ = x[eofInComment-14]
	_ = x[eofInDoctype-15]
	_ = x[eofInScriptHTMLCommentLikeText-16]
	_ = x[eofInTag-17]
	_ = x[incorrectlyClosedComment-18]
	_ = x[incorrectlyOpenedComment-19]
	_ = x[invalidCharacterSequenceAfterDoctypeName-20]
	_ = x[invalidFirstCharacterOfTagName-21]
	_ = x[missingAttributeValue-22]
	_ = x[missingDoctypeName-23]
	_ = x[missingDoctypePublicIdentifier-24]
	_ = x[missingDoctypeSystemIdentifier-25]
	_ = x[missingEndTagName-26]
	_ = x[missingQuoteBeforeDoctypePublicIdentifier-27]
	_ = x[missingQuoteBeforeDoctypeSystemIdentifier-28]
	_ = x[missingSemiColonAfterCharacterReference-29]
	_ = x[missingWhitespaceAfterDoctypePublicKeyword-30]
	_ = x[missingWhitespaceAfterDoctypeSystemKeyword-31]
	_ = x[missingWhitespaceBeforeDoctypeName-32]
	_ = x[missingWhitespaceBetweenAttributes-33]
	_ = x[missingWhitespaceBetweenDoctypePublicAndSystemIdentifiers-34]
	_ = x[nestedComment-35]
	_ = x[noncharacterCharacterReference-36]
	_ = x[noncharacterInInputStream-37]
	_ = x[nonVoidHTMLElementStartTagWithTrailingSolidus-38]
	_ = x[nullCharacterReference-39]
	_ = x[surrogateCharacterReference-40]
	_ = x[surrogateInInputStream-41]
	_ = x[unexpectedCharacterAfterDoctypeSystemIdentifier-42]
	_ = x[unexpectedCharacterInAttributeName-43]
	_ = x[unexpectedCharacterInUnquotedAttributeValue-44]
	_ = x[unexpectedEqualsSignBeforeAttributeName-45]
	_ = x[unexpectedNullCharacter-46]
	_ = x[unexpectedQuestionMakrInsteadofTagName-47]
	_ = x[unexpectedSolidusInTag-48]
	_ = x[unknownNamedCharacterReference-49]
	_ = x[generalParseError-50]
}

const _parseError_name = "noErrorabruptClosingOfEmptyCommentabruptDoctypePublicIdentifierabruptDoctypeSystemIdentifierabsenceOfDigitsInNumericCharacterReferencecdataInHTMLContentcharacterReferenceOutsideUnicodeRangecontrolCharacterInInputSteamcontrolCharacterReferenceendTagWithAttributesduplicateAttributeendTagWithTrailingSoliduseofBeforeTagNameeofInCdataeofInCommenteofInDoctypeeofInScriptHTMLCommentLikeTexteofInTagincorrectlyClosedCommentincorrectlyOpenedCommentinvalidCharacterSequenceAfterDoctypeNameinvalidFirstCharacterOfTagNamemissingAttributeValuemissingDoctypeNamemissingDoctypePublicIdentifiermissingDoctypeSystemIdentifiermissingEndTagNamemissingQuoteBeforeDoctypePublicIdentifiermissingQuoteBeforeDoctypeSystemIdentifiermissingSemiColonAfterCharacterReferencemissingWhitespaceAfterDoctypePublicKeywordmissingWhitespaceAfterDoctypeSystemKeywordmissingWhitespaceBeforeDoctypeNamemissingWhitespaceBetweenAttributesmissingWhitespaceBetweenDoctypePublicAndSystemIdentifiersnestedCommentnoncharacterCharacterReferencenoncharacterInInputStreamnonVoidHTMLElementStartTagWithTrailingSolidusnullCharacterReferencesurrogateCharacterReferencesurrogateInInputStreamunexpectedCharacterAfterDoctypeSystemIdentifierunexpectedCharacterInAttributeNameunexpectedCharacterInUnquotedAttributeValueunexpectedEqualsSignBeforeAttributeNameunexpectedNullCharacterunexpectedQuestionMakrInsteadofTagNameunexpectedSolidusInTagunknownNamedCharacterReferencegeneralParseError"

var _parseError_index = [...]uint16{0, 7, 34, 63, 92, 134, 152, 189, 217, 242, 262, 280, 305, 321, 331, 343, 355, 385, 393, 417, 441, 481, 511, 532, 550, 580, 610, 627, 668, 709, 748, 790, 832, 866, 900, 957, 970, 1000, 1025, 1070, 1092, 1119, 1141, 1188, 1222, 1265, 1304, 1327, 1365, 1387, 1417, 1434}

func (i parseError) String() string {
	if i >= parseError(len(_parseError_index)-1) {
		return "parseError(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _parseError_name[_parseError_index[i]:_parseError_index[i+1]]
}
