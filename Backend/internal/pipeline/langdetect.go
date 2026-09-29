package pipeline

import (
	"strings"
	"unicode"
)

// DetectLanguage applies a lightweight heuristic language detector so the
// pipeline can route content without calling an external service.
//
// Strategy:
//  1. Script detection for non-latin scripts (Devanagari → hi, Bengali → bn,
//     Tamil → ta, Telugu → te, Arabic → ar, Cyrillic → ru, Han → zh,
//     Hangul → ko).
//  2. Stop-word lexicon voting for latin-script languages (en, es, pt, fr,
//     id + Hinglish markers).
//
// It returns a short language code and a confidence in [0,1].
func DetectLanguage(text string) (string, float64) {
	if strings.TrimSpace(text) == "" {
		return "und", 0
	}

	var (
		letters  int
		scriptHi int
		scriptBn int
		scriptTa int
		scriptTe int
		scriptAr int
		scriptRu int
		scriptZh int
		scriptKo int
	)
	for _, r := range text {
		if unicode.IsLetter(r) {
			letters++
		}
		switch {
		case unicode.In(r, unicode.Devanagari):
			scriptHi++
		case unicode.In(r, unicode.Bengali):
			scriptBn++
		case unicode.In(r, unicode.Tamil):
			scriptTa++
		case unicode.In(r, unicode.Telugu):
			scriptTe++
		case unicode.In(r, unicode.Arabic):
			scriptAr++
		case unicode.In(r, unicode.Cyrillic):
			scriptRu++
		case unicode.In(r, unicode.Han):
			scriptZh++
		case unicode.In(r, unicode.Hangul):
			scriptKo++
		}
	}
	if letters == 0 {
		return "und", 0
	}

	type scriptCount struct {
		code  string
		count int
	}
	scripts := []scriptCount{
		{"hi", scriptHi},
		{"bn", scriptBn},
		{"ta", scriptTa},
		{"te", scriptTe},
		{"ar", scriptAr},
		{"ru", scriptRu},
		{"zh", scriptZh},
		{"ko", scriptKo},
	}
	best := scriptCount{}
	for _, s := range scripts {
		if s.count > best.count {
			best = s
		}
	}
	if best.count > 0 {
		ratio := float64(best.count) / float64(letters)
		if ratio >= 0.15 {
			conf := ratio * 3
			if conf > 0.99 {
				conf = 0.99
			}
			if conf < 0.5 {
				conf = 0.5
			}
			return best.code, conf
		}
	}

	// Latin path: stop-word lexicon voting.
	words := tokenizeLatin(text)
	if len(words) == 0 {
		return "und", 0.1
	}
	hits := map[string]int{}
	for _, w := range words {
		if lang, ok := stopWords[w]; ok {
			hits[lang]++
		}
	}
	bestLang, bestHits := "", 0
	for lang, n := range hits {
		if n > bestHits {
			bestLang, bestHits = lang, n
		}
	}
	if bestHits >= 2 {
		conf := float64(bestHits) / float64(len(words)) * 4
		if conf > 0.95 {
			conf = 0.95
		}
		if conf < 0.4 {
			conf = 0.4
		}
		return bestLang, conf
	}
	if bestHits == 1 {
		return bestLang, 0.4
	}
	if len(words) >= 5 {
		return "en", 0.35 // default assumption for unmarked latin text
	}
	return "und", 0.2
}

func tokenizeLatin(text string) []string {
	lower := strings.ToLower(text)
	fields := strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if len([]rune(f)) >= 2 {
			out = append(out, f)
		}
	}
	return out
}

// stopWords maps frequent function words to their language. Deliberately
// small, high-precision sets — this runs inline in the pipeline.
var stopWords = map[string]string{
	// English
	"the": "en", "and": "en", "is": "en", "of": "en", "to": "en", "in": "en",
	"it": "en", "for": "en", "on": "en", "with": "en", "this": "en",
	"that": "en", "are": "en", "was": "en", "be": "en", "as": "en", "at": "en",
	"by": "en", "from": "en", "or": "en", "an": "en", "we": "en", "you": "en",
	"they": "en", "not": "en", "have": "en", "has": "en", "but": "en",
	// Hinglish / Hindi romanized
	"hai": "hi", "hain": "hi", "kya": "hi", "aur": "hi", "nahi": "hi",
	"mein": "hi", "ke": "hi", "ki": "hi", "ka": "hi", "se": "hi", "par": "hi",
	"yeh": "hi", "wo": "hi", "kar": "hi", "koi": "hi", "bhi": "hi", "tha": "hi",
	"raha": "hi", "rahe": "hi", "hoga": "hi", "liye": "hi", "bahut": "hi",
	// Spanish
	"el": "es", "la": "es", "los": "es", "las": "es", "de": "es", "que": "es",
	"y": "es", "en": "es", "un": "es", "una": "es", "es": "es", "por": "es",
	"con": "es", "para": "es", "no": "es", "del": "es", "al": "es",
	// Portuguese
	"os": "pt", "do": "pt", "da": "pt", "em": "pt", "um": "pt",
	"uma": "pt", "com": "pt", "nao": "pt", "não": "pt", "mais": "pt",
	// French
	"le": "fr", "les": "fr", "des": "fr", "du": "fr", "et": "fr", "une": "fr",
	"est": "fr", "pour": "fr", "dans": "fr", "ce": "fr", "au": "fr",
	"aux": "fr", "sur": "fr",
	// Indonesian
	"yang": "id", "dan": "id", "di": "id", "dari": "id",
	"untuk": "id", "dengan": "id", "ini": "id", "itu": "id", "tidak": "id",
	"ada": "id",
}
