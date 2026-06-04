package gemini

import "google.golang.org/genai"

// Configure safety settings to block none
var blockNone = genai.HarmBlockThresholdBlockNone
var safetySettings = []*genai.SafetySetting{
	{Category: genai.HarmCategoryHarassment, Threshold: blockNone},
	{Category: genai.HarmCategoryHateSpeech, Threshold: blockNone},
	{Category: genai.HarmCategorySexuallyExplicit, Threshold: blockNone},
	{Category: genai.HarmCategoryDangerousContent, Threshold: blockNone},
}
