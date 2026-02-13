package yt

import (
	"regexp"
	"strings"
	"unicode"
)

var bracketedContentRegex = regexp.MustCompile(`[\(\[].*?[\)\]]`)
var urlRegex = regexp.MustCompile(`http\S+`)
var emailRegex = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)

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
	'"':      true, // U+0022 straight double quote (quotation Mark)
	'\'':     true, // U+0027 straight single quote (apostrophe)
	'\u201C': true, // U+201C left double quotation mark
	'\u201D': true, // U+201D right double quotation mark
	'\u2018': true, // U+2018 left single quotation mark
	'\u2019': true, // U+2019 right single quotation mark
	'\u0060': true, // U+0060 grave accent
	'\u00B4': true, // U+00B4 acute accent
}

var wierdSingleQuotes = map[rune]bool{
	'\u2018': true, // U+2018 left single quotation mark
	'\u2019': true, // U+2019 right single quotation mark
	'\u0060': true, // U+0060 grave accent
	'\u00B4': true, // U+00B4 acute accent
}

var wierdDoubleQuotes = map[rune]bool{
	'\u201C': true, // U+201C left double quotation mark
	'\u201D': true, // U+201D right double quotation mark
}

// Normalize the YouTube video title
func normalizeTitle(title string, cutOffs []string) string {

	// Cut off the title at certain substrings
	for _, substring := range cutOffs {
		title = strings.Split(title, substring)[0]
	}

	// Remove bracketed content
	title = bracketedContentRegex.ReplaceAllString(title, "")

	// Remove leading/trailing white space and split on empty space(s)
	words := strings.Fields(title)

	// Inspect/mutate/exclude the words
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

		var firstQuote string
		var lastQuote string

		// Remove quotation marks from the word at start/end
		// if any and store them for later use
		if len(runes) > 1 {
			if quotes[runes[0]] {

				// Use straight quotes
				if wierdSingleQuotes[runes[0]] {
					runes[0] = '\''
				} else if wierdDoubleQuotes[runes[0]] {
					runes[0] = '"'
				}

				firstQuote = string(runes[0])
				runes = runes[1:]
			}

			lastIndex := len(runes) - 1
			if quotes[runes[lastIndex]] {

				// Use straight quotes
				if wierdSingleQuotes[runes[lastIndex]] {
					runes[lastIndex] = '\''
				} else if wierdDoubleQuotes[runes[lastIndex]] {
					runes[lastIndex] = '"'
				}

				lastQuote = string(runes[lastIndex])
				runes = runes[:lastIndex]
			}
		}

		// Replace weird single quotes inside the word
		// with straight single quote (apostrophe)
		for i, r := range runes {
			if wierdSingleQuotes[r] {
				runes[i] = '\''
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
		// So it should be all lowercase.
		if i > 0 && preps[currentWord] && !puncts[previousWordLastRune] {
			word = firstQuote + currentWord + lastQuote

			// Capitalize the word
		} else {
			firstLetter := string(unicode.ToUpper(runes[0]))
			word = firstQuote + firstLetter + string(runes[1:]) + lastQuote
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

// normalizeDescription removes URLs and emails from a text
func normalizeDescription(text string) string {
	text = urlRegex.ReplaceAllString(text, "")
	return emailRegex.ReplaceAllString(text, "")
}
