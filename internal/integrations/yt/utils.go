package yt

import "regexp"

var bracketedContent = regexp.MustCompile(`[\(\[].*?[\)\]]`)
var extraSpace = regexp.MustCompile(`\s+`)

// Common prepositions
var preps = map[string]bool{
	"at":  true,
	"by":  true,
	"for": true,
	"in":  true,
	"of":  true,
	"off": true,
	"the": true,
	"and": true,
	"or":  true,
	"nor": true,
	"a":   true,
	"an":  true,
	"on":  true,
	"out": true,
	"to":  true,
	"up":  true,
	"as":  true,
	"but": true,
	"per": true,
	"via": true,
	"vs":  true,
	"vs.": true,
}

// Punctuations
var puncts = map[rune]bool{
	':':      true,
	'.':      true,
	'!':      true,
	'?':      true,
	'\u002D': true, // U+002D hyphen-minus
	'\u2014': true, // U+2014 em dash
	'\u2013': true, // U+2013 en dash
	'|':      true,
}

// Other weird strings
var others = []string{"//", "--"}

// Usual quotes
var quotes = map[rune]bool{
	'"':      true, // U+0022 straight double quote
	'\'':     true, // U+0027 straight single quote
	'\u201C': true, // U+201C left double quotation mark
	'\u201D': true, // U+201D right double quotation mark
	'\u2018': true, // U+2018 left single quotation mark
	'\u2019': true, // U+2019 right single quotation mark
}
