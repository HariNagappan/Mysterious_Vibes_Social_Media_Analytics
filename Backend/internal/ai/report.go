// Package ai produces the PulseGraph analytical report. Every number in a
// report must be grounded in already-computed metrics — the summarizer never
// invents facts. That guarantee is enforced by ValidateReport.
package ai

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Report is a generated analytical narrative.
type Report struct {
	Title            string    `json:"title"`
	ExecutiveSummary string    `json:"executive_summary"`
	KeyFindings      []string  `json:"key_findings"`
	Sections         []Section `json:"sections"`
	GeneratedAt      time.Time `json:"generated_at"`
	ModelVersion     string    `json:"model_version"`
	Mode             string    `json:"mode"`
}

// Section is one titled block of the report body.
type Section struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

// Text flattens the report for guardrail checks.
func (r *Report) Text() string {
	var b strings.Builder
	b.WriteString(r.Title)
	b.WriteString("\n")
	b.WriteString(r.ExecutiveSummary)
	b.WriteString("\n")
	for _, finding := range r.KeyFindings {
		b.WriteString(finding)
		b.WriteString("\n")
	}
	for _, section := range r.Sections {
		b.WriteString(section.Heading)
		b.WriteString("\n")
		b.WriteString(section.Body)
		b.WriteString("\n")
	}
	return b.String()
}

var numberRe = regexp.MustCompile(`-?\d+(?:\.\d+)?`)

// UnexplainedNumbers returns numeric claims in the text that are not grounded
// in the source metrics.
//
// A value is considered grounded when it equals (within 1e-3) a source value,
// a rounded integer of one, a percentage form (×100), a share form (÷100), or
// falls in the structural allowlist: years 2000–2100 and small integers 0–12
// (top-N lists, ordinal references).
func UnexplainedNumbers(text string, source map[string]float64) []float64 {
	var out []float64
	for _, match := range numberRe.FindAllString(text, -1) {
		value, err := strconv.ParseFloat(match, 64)
		if err != nil {
			continue
		}
		if numberExplained(value, source) {
			continue
		}
		out = append(out, value)
	}
	return out
}

func numberExplained(value float64, source map[string]float64) bool {
	if value == math.Trunc(value) && value >= 2000 && value <= 2100 {
		return true
	}
	if value == math.Trunc(value) && value >= 0 && value <= 12 {
		return true
	}
	for _, candidate := range source {
		if almostEqual(value, candidate) ||
			almostEqual(value, candidate*100) ||
			almostEqual(value, candidate/100) ||
			almostEqual(value, math.Round(candidate)) ||
			almostEqual(value, math.Round(candidate*100)/100) {
			return true
		}
	}
	return false
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 1e-3
}

// ValidateReport returns an error when the report makes ungrounded numeric
// claims. The ML-service path runs this check before serving a report.
func ValidateReport(report *Report, source map[string]float64) error {
	if unexplained := UnexplainedNumbers(report.Text(), source); len(unexplained) > 0 {
		return fmt.Errorf("ai: report contains %d ungrounded numeric claim(s), e.g. %.2f", len(unexplained), unexplained[0])
	}
	return nil
}

// Round is the shared rounding helper for deterministic inputs.
func Round(v float64, places int) float64 {
	factor := math.Pow10(places)
	return math.Round(v*factor) / factor
}
