package ingestion

import "testing"

func TestNormalizeRequiresFields(t *testing.T) {
	if _, err := Normalize(RawPost{}, 1); err == nil {
		t.Fatal("expected error for an empty raw post")
	}
	if _, err := Normalize(RawPost{Platform: "twitter"}, 1); err == nil {
		t.Fatal("expected error for a missing external id")
	}
	if _, err := Normalize(RawPost{Platform: "twitter", ExternalID: "tw-1"}, 1); err == nil {
		t.Fatal("expected error for a missing author reference")
	}
}

func TestNormalizeDetectsLanguage(t *testing.T) {
	normalized, err := Normalize(RawPost{
		Platform:   "Twitter",
		ExternalID: "tw-1",
		AuthorRef:  "x:user:a",
		Content:    "The flood rescue effort continues today with volunteers arriving from nearby towns",
	}, 7)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if normalized.Platform != "twitter" {
		t.Fatalf("platform should be lowercased, got %s", normalized.Platform)
	}
	if normalized.Language != "en" {
		t.Fatalf("expected en, got %s", normalized.Language)
	}
	if normalized.TopicID != 7 {
		t.Fatalf("topic id not propagated: %d", normalized.TopicID)
	}
}

func TestNormalizeMentionsDeduped(t *testing.T) {
	normalized, err := Normalize(RawPost{
		Platform:    "twitter",
		ExternalID:  "tw-2",
		AuthorRef:   "x:user:a",
		Content:     "hello @b and @b again",
		MentionRefs: []string{"x:user:b", "x:user:b", "x:user:a", "", "x:user:c"},
	}, 1)
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(normalized.MentionRefs) != 2 {
		t.Fatalf("expected deduped mentions without self, got %v", normalized.MentionRefs)
	}
	if normalized.MentionRefs[0] != "x:user:b" || normalized.MentionRefs[1] != "x:user:c" {
		t.Fatalf("unexpected mentions: %v", normalized.MentionRefs)
	}
}
