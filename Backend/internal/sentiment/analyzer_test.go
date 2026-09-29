package sentiment

import "testing"

func TestAnalyzeLexiconPositive(t *testing.T) {
	res := AnalyzeLexicon("This is a great and wonderful relief, so happy with the rescue team. Best news today!")
	if res.Score <= 0 {
		t.Fatalf("expected positive score, got %f", res.Score)
	}
	if res.ModelVersion != FallbackModelVersion {
		t.Fatalf("expected fallback model version, got %s", res.ModelVersion)
	}
	if res.Confidence <= 0 || res.Confidence > 1 {
		t.Fatalf("confidence out of range: %f", res.Confidence)
	}
}

func TestAnalyzeLexiconNegative(t *testing.T) {
	res := AnalyzeLexicon("Furious about this shameless corruption. The situation is a tragedy and victims are suffering.")
	if res.Score >= 0 {
		t.Fatalf("expected negative score, got %f", res.Score)
	}
}

func TestAnalyzeLexiconNegationFlips(t *testing.T) {
	positive := AnalyzeLexicon("the response was good")
	negated := AnalyzeLexicon("the response was not good")
	if !(negated.Score < positive.Score) {
		t.Fatalf("negation should reduce the score: %f vs %f", negated.Score, positive.Score)
	}
}

func TestAnalyzeLexiconHinglish(t *testing.T) {
	res := AnalyzeLexicon("yeh bahut kharab hai, koi madad nahi mil rahi")
	if res.Score >= 0 {
		t.Fatalf("expected negative hinglish score, got %f", res.Score)
	}
}

func TestAnalyzeLexiconNeutralFallback(t *testing.T) {
	res := AnalyzeLexicon("")
	if res.Emotion != "neutral" || res.Score != 0 {
		t.Fatalf("expected neutral result, got %+v", res)
	}
}
