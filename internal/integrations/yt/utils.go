package yt

import (
	"regexp"
	"strings"
)

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

// Normalize the YouTube video title
func normalizeTitle(title string) string {

	// Cut off the title at certain substrings
	for _, substring := range []string{" I SLICE ", " // ", " | "} {
		title = strings.Split(title, substring)[0]
	}

	// Remove bracketed content
	title = strings.TrimSpace(bracketedContent.ReplaceAllString(title, ""))

	// Remove extra spaces
	title = extraSpace.ReplaceAllString(title, " ")

	// Split the title into words and remove the last word if it's 'documentary'
	words := strings.Split(title, " ")
	if strings.ToLower(words[len(words)-1]) == "documentary" {
		words = words[1:]
	}

	// Iterate the words and mutate them
	for i, w := range words {
		// Convert word to runes
		runes := []rune(w)

		var fq string
		var lq string

		// Remove quotation marks from the word at start/end
		// and store them for later use
		if len(runes) > 1 {
			if quotes[runes[0]] {
				fq = string(runes[0])
				w = string(runes[1:])
			}

			if quotes[runes[len(runes)-1]] {
				lq = string(runes[len(runes)-1])
				w = string(runes[:1])
			}
		}

		// If not the first word try to lowercase the word
		if i > 0 {
			currentWord := strings.ToLower(w)
			previousWord := []rune(words[i-1])
			lastRune := previousWord[len(previousWord)-1]
			// The word is a preposition but not after a punctuation
			if preps[currentWord] && !puncts[lastRune] {
				// Replace the actual word in the slice
				words[i] = fq + currentWord + lq
			}
		}

	}

	return strings.Join(words, " ")
}
