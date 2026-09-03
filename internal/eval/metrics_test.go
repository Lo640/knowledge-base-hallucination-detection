package eval

import (
	"testing"

	"github.com/example/knowledge-base-hallucination-detection/internal/model"
)

func TestCalculateAndSplitMismatches(t *testing.T) {
	results := []model.SampleResult{
		{ID: "tp", Prediction: model.Detection{Hallucinate: true}, GroundTruth: model.GroundTruth{Hallucinate: true}},
		{ID: "tn", Prediction: model.Detection{Hallucinate: false}, GroundTruth: model.GroundTruth{Hallucinate: false}},
		{ID: "fp", Prediction: model.Detection{Hallucinate: true}, GroundTruth: model.GroundTruth{Hallucinate: false}},
		{ID: "fn", Prediction: model.Detection{Hallucinate: false}, GroundTruth: model.GroundTruth{Hallucinate: true}},
	}
	m := Calculate(results)
	if m.TP != 1 || m.TN != 1 || m.FP != 1 || m.FN != 1 {
		t.Fatalf("unexpected confusion matrix: %+v", m)
	}
	if m.Precision != 0.5 || m.Recall != 0.5 || m.F1 != 0.5 {
		t.Fatalf("unexpected metrics: %+v", m)
	}
	fn, fp := SplitMismatches(results)
	if len(fn) != 1 || fn[0].ID != "fn" || len(fp) != 1 || fp[0].ID != "fp" {
		t.Fatalf("unexpected mismatch split: fn=%v fp=%v", fn, fp)
	}
}
