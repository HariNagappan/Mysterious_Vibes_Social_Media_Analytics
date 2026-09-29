package sentiment

import (
	"math"
	"strings"
	"unicode"
)

// FallbackModelVersion labels results produced by the in-process lexicon
// analyzer (used when the Python ML service is unreachable, so the pipeline
// degrades gracefully instead of stalling).
const FallbackModelVersion = "lexicon-v1"

// AnalyzeLexicon is a deterministic English + Hinglish lexicon scorer. It is
// intentionally transparent: familiar word lists, negation handling and
// intensity modifiers — good enough to keep the demo pipeline alive.
func AnalyzeLexicon(text string) Result {
	tokens := tokenize(text)
	if len(tokens) == 0 {
		return Result{Emotion: "neutral", Score: 0, Confidence: 0.5, ModelVersion: FallbackModelVersion}
	}

	categories := map[string]int{}
	var positive, negative float64
	var hits int

	for i, tok := range tokens {
		word := normalizeToken(tok)
		if word == "" {
			continue
		}

		weight := 1.0
		if hasIntensifier(tokens, i) {
			weight = 1.5
		}
		negated := isNegated(tokens, i)

		if cat, ok := positiveWords[word]; ok {
			hits++
			categories[cat]++
			if negated {
				negative += weight
			} else {
				positive += weight
			}
			continue
		}
		if cat, ok := negativeWords[word]; ok {
			hits++
			categories[cat]++
			if negated {
				positive += weight * 0.5
			} else {
				negative += weight
			}
		}
	}

	score := clamp((positive-negative)/4.0, -1, 1)

	emotion := "neutral"
	best := 0
	for cat, count := range categories {
		if count > best {
			best = count
			emotion = cat
		}
	}
	if emotion == "neutral" {
		if score > 0.15 {
			emotion = "joy"
		} else if score < -0.15 {
			emotion = "anger"
		}
	}

	confidence := 0.5
	if hits > 0 {
		confidence = clamp(0.55+0.08*float64(hits), 0.55, 0.92)
	}

	return Result{Emotion: emotion, Score: score, Confidence: confidence, ModelVersion: FallbackModelVersion}
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func normalizeToken(tok string) string {
	return strings.ToLower(strings.TrimSpace(tok))
}

var negators = map[string]struct{}{
	"not": {}, "no": {}, "never": {}, "none": {}, "cannot": {}, "cant": {},
	"can't": {}, "dont": {}, "don't": {}, "isnt": {}, "isn't": {}, "wasnt": {},
	"wasn't": {}, "wont": {}, "won't": {}, "nahi": {}, "nahin": {}, "na": {},
	"mat": {}, "bilkul": {},
}

var intensifiers = map[string]struct{}{
	"very": {}, "extremely": {}, "really": {}, "so": {}, "totally": {},
	"bahut": {}, "बहुत": {}, "बहोत": {}, "hadd": {}, "बेहद": {},
}

// isNegated reports whether one of the two preceding tokens negates the word
// at index i.
func isNegated(tokens []string, i int) bool {
	for back := 1; back <= 2; back++ {
		j := i - back
		if j < 0 {
			break
		}
		if _, ok := negators[normalizeToken(tokens[j])]; ok {
			return true
		}
	}
	return false
}

// hasIntensifier reports whether the token right before i intensifies it.
func hasIntensifier(tokens []string, i int) bool {
	j := i - 1
	if j < 0 {
		return false
	}
	_, ok := intensifiers[normalizeToken(tokens[j])]
	return ok
}

func clamp(v, lo, hi float64) float64 {
	return math.Min(hi, math.Max(lo, v))
}

// positiveWords maps sentiment words onto their emotion category.
var positiveWords = map[string]string{
	// joy (English)
	"good": "joy", "great": "joy", "awesome": "joy", "amazing": "joy",
	"fantastic": "joy", "love": "joy", "happy": "joy", "glad": "joy",
	"relief": "joy", "relieved": "joy", "success": "joy", "win": "joy",
	"wins": "joy", "winning": "joy", "brilliant": "joy", "excellent": "joy",
	"proud": "joy", "hope": "joy", "hopeful": "joy", "support": "joy",
	"helpful": "joy", "thanks": "joy", "thank": "joy", "best": "joy",
	"better": "joy", "safe": "joy", "rescued": "joy", "recovered": "joy",
	"calm": "joy", "brave": "joy", "courage": "joy", "generous": "joy",
	"kind": "joy", "peace": "joy",
	// surprise (positive-leaning)
	"surprising": "surprise", "unexpectedly": "surprise", "unbelievable": "surprise",
	// joy (Hinglish)
	"accha": "joy", "achha": "joy", "badhiya": "joy", "shandar": "joy",
	"mubarak": "joy", "khushi": "joy", "jeet": "joy", "safal": "joy",
	"sahi": "joy", "madad": "joy",
}

// negativeWords maps sentiment words onto their emotion category.
var negativeWords = map[string]string{
	// anger
	"angry": "anger", "outrage": "anger", "outrageous": "anger",
	"corrupt": "anger", "corruption": "anger", "scam": "anger",
	"shame": "anger", "shameless": "anger", "furious": "anger", "hate": "anger",
	"disgrace": "anger", "betrayal": "anger", "liar": "anger", "lies": "anger",
	"propaganda": "anger", "fraud": "anger", "cheating": "anger",
	"cheat": "anger", "stupid": "anger", "pathetic": "anger",
	"incompetent": "anger", "jhooth": "anger", "jhuth": "anger",
	"dhoka": "anger", "galat": "anger", "bakwas": "anger",
	// fear
	"fear": "fear", "feared": "fear", "panic": "fear", "scared": "fear",
	"danger": "fear", "dangerous": "fear", "threat": "fear",
	"warning": "fear", "worried": "fear", "worry": "fear", "anxiety": "fear",
	"crisis": "fear", "emergency": "fear", "terrified": "fear",
	"alarming": "fear", "urgent": "fear", "khatra": "fear", "dar": "fear",
	"darr": "fear", "pareshani": "fear", "mushkil": "fear",
	// sadness
	"sad": "sadness", "loss": "sadness", "lost": "sadness", "died": "sadness",
	"dead": "sadness", "death": "sadness", "victims": "sadness",
	"victim": "sadness", "tragic": "sadness", "tragedy": "sadness",
	"grief": "sadness", "missing": "sadness", "broken": "sadness",
	"sorrow": "sadness", "mourning": "sadness", "devastated": "sadness",
	"dukhi": "sadness", "nuksan": "sadness",
	// disgust
	"disgusting": "disgust", "gross": "disgust", "vile": "disgust",
	"nasty": "disgust", "repulsive": "disgust", "filth": "disgust",
	"filthy": "disgust", "shameful": "disgust", "revolting": "disgust",
	"kharab": "disgust", "bura": "disgust", "bekar": "disgust",
	// surprise (negative-leaning)
	"shocking": "surprise", "shocked": "surprise", "stunned": "surprise",
	"astonishing": "surprise", "stunning": "surprise",
}
