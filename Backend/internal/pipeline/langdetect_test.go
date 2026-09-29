package pipeline

import "testing"

func TestDetectLanguageDevanagari(t *testing.T) {
	lang, conf := DetectLanguage("बाढ़ की स्थिति बहुत गंभीर है, राहत की जरूरत है")
	if lang != "hi" {
		t.Fatalf("expected hi, got %s", lang)
	}
	if conf <= 0 {
		t.Fatal("expected positive confidence for script detection")
	}
}

func TestDetectLanguageEnglish(t *testing.T) {
	lang, _ := DetectLanguage("The flood situation is getting worse and the local authorities are not responding to this emergency")
	if lang != "en" {
		t.Fatalf("expected en, got %s", lang)
	}
}

func TestDetectLanguageHinglish(t *testing.T) {
	lang, _ := DetectLanguage("yeh situation bahut kharab hai, koi madad nahi kar raha")
	if lang != "hi" {
		t.Fatalf("expected hi (hinglish markers), got %s", lang)
	}
}

func TestDetectLanguageEmpty(t *testing.T) {
	lang, conf := DetectLanguage("   ")
	if lang != "und" || conf != 0 {
		t.Fatalf("expected und/0 for empty text, got %s/%f", lang, conf)
	}
}
