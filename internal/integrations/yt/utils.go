package yt

import (
	"regexp"
	"strings"
	"unicode"
)

var bracketedContent = regexp.MustCompile(`[\(\[].*?[\)\]]`)
var extraSpace = regexp.MustCompile(`\s+`)
var urls = regexp.MustCompile(`http\S+`)

var unwantedWords = map[string]bool{
	"documentary": true,
	"4k":          true,
}

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

	// Split the title into words and inspect/mutate/exclude them
	words := strings.Split(title, " ")
	var result []string
	for i, word := range words {

		// Exclude unwanted words
		if unwantedWords[strings.ToLower(word)] {
			continue
		}

		// Convert word to runes slice
		runes := []rune(word)

		// Exclude hashtag words
		if runes[0] == '#' {
			continue
		}

		// First and last quote
		var fq string
		var lq string

		// Remove quotation marks from the word at start/end
		// if any and store them for later use
		if len(runes) > 1 {
			if quotes[runes[0]] {
				fq = string(runes[0])
				runes = runes[1:]
			}

			lastIndex := len(runes) - 1
			if quotes[runes[lastIndex]] {
				lq = string(runes[lastIndex])
				runes = runes[:lastIndex]
			}
		}

		// Loweracse the resulting current word
		currentWord := strings.ToLower(string(runes))

		// Take the last rune from the previous word
		var previousWordLastRune rune
		if i > 0 {
			previousWord := []rune(words[i-1])
			previousWordLastRune = previousWord[len(previousWord)-1]
		}

		// This is not the first word.
		// The word is a preposition but not after a punctuation.
		// So it should be lowercase, not capitalized.
		if i > 0 && preps[currentWord] && !puncts[previousWordLastRune] {
			word = string(fq) + currentWord + string(lq)

			// Capitalize the word if not already capitalized
		} else if !unicode.IsUpper(runes[0]) {
			word = string(fq) +
				string(unicode.ToUpper(runes[0])) +
				string(runes[1:]) +
				string(lq)
		}

		result = append(result, word)
	}

	return strings.Join(result, " ")
}

// Normalize tags, remove duplicate words
func normalizeTags(tags []string, title, description string) (result string) {

	// Assemble a map
	seen := map[string]bool{
		"documentary":   true,
		"documentaries": true,
	}

	// Make title and description lowercase and
	// split them in words on non-alphanumeric runes
	used := strings.FieldsFunc(
		strings.ToLower(title+" "+description), func(c rune) bool {
			return !unicode.IsLetter(c) && !unicode.IsNumber(c)
		})

	// Put those words in the seen map
	for _, word := range used {
		seen[word] = true
	}

	// Go over the tags and see what you can include
	for _, tag := range tags {
		lowerTag := strings.ToLower(tag)
		if !seen[lowerTag] {
			seen[lowerTag] = true
			result += tag + " "
		}
	}

	return strings.TrimSpace(result)
}
