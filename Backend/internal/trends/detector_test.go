package trends

import "testing"

func TestDetectRisingWithSpike(t *testing.T) {
	buckets := []Bucket{
		{Count: 1}, {Count: 1}, {Count: 2},
		{Count: 3}, {Count: 4}, {Count: 20},
	}
	det := Detect(buckets)
	if det.Growth <= 0 {
		t.Fatalf("expected positive growth, got %f", det.Growth)
	}
	if det.Velocity <= 0 {
		t.Fatalf("expected positive velocity, got %f", det.Velocity)
	}
	if !det.Spike {
		t.Fatal("expected spike detection on the final bucket")
	}
}

func TestDetectFlatSeries(t *testing.T) {
	buckets := []Bucket{{Count: 5}, {Count: 5}, {Count: 5}, {Count: 5}, {Count: 5}}
	det := Detect(buckets)
	if det.Growth != 0 || det.Velocity != 0 || det.Spike {
		t.Fatalf("flat series should not show growth/spike: %+v", det)
	}
}

func TestExtractKeywordsFiltersStopwords(t *testing.T) {
	keywords := ExtractKeywords([]string{
		"the flood is bad and the flood is dangerous",
		"flood water rising near the relief camp",
	}, 5)
	if len(keywords) == 0 {
		t.Fatal("expected keywords")
	}
	if keywords[0].Keyword != "flood" {
		t.Fatalf("expected flood to rank first, got %+v", keywords)
	}
}

func TestScoreBounds(t *testing.T) {
	if score := Score(-5, -5, 0); score != 0 {
		t.Fatalf("negative inputs should score 0, got %f", score)
	}
	if score := Score(10, 100, 100000); score <= 0 || score > 100 {
		t.Fatalf("score out of range: %f", score)
	}
}

func TestBucketize(t *testing.T) {
	series := Bucketize([]int64{100, 149, 150, 999}, 100, 3, 50)
	if len(series) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(series))
	}
	if series[0].Count != 2 || series[1].Count != 1 || series[2].Count != 1 {
		t.Fatalf("unexpected bucket counts: %+v", series)
	}
}
