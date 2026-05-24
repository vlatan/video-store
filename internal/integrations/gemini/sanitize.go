package gemini

import "strings"

// sanitizePrompt is replacing visceral/graphic verbs and nouns with synonyms
func SanitizePrompt(input string) string {

	replacer := strings.NewReplacer(
		"execution", "killing",
		"beheaded", "killed",
		"beheading", "killing",
		"slaughtered", "attacked",
		"massacre", "incident",
		"genocide", "conflict",

		"rape", "abuse",
		"raped", "abused",
		"sex slave", "captive",
		"sexual slavery", "forced captivity",
	)

	return replacer.Replace(input)
}
