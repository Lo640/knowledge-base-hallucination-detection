package eval

import (
	"sort"

	"github.com/example/knowledge-base-hallucination-detection/internal/model"
)

func Calculate(results []model.SampleResult) model.Metrics {
	m := model.Metrics{Total: len(results)}
	for _, result := range results {
		switch {
		case result.Prediction.Hallucinate && result.GroundTruth.Hallucinate:
			m.TP++
		case !result.Prediction.Hallucinate && !result.GroundTruth.Hallucinate:
			m.TN++
		case result.Prediction.Hallucinate:
			m.FP++
		default:
			m.FN++
		}
	}
	if m.TP+m.FP > 0 {
		m.Precision = float64(m.TP) / float64(m.TP+m.FP)
	}
	if m.TP+m.FN > 0 {
		m.Recall = float64(m.TP) / float64(m.TP+m.FN)
	}
	if m.Precision+m.Recall > 0 {
		m.F1 = 2 * m.Precision * m.Recall / (m.Precision + m.Recall)
	}
	return m
}

func SplitMismatches(results []model.SampleResult) (falseNegatives, falsePositives []model.SampleResult) {
	for _, result := range results {
		if result.GroundTruth.Hallucinate && !result.Prediction.Hallucinate {
			falseNegatives = append(falseNegatives, result)
		}
		if !result.GroundTruth.Hallucinate && result.Prediction.Hallucinate {
			falsePositives = append(falsePositives, result)
		}
	}
	sort.Slice(falseNegatives, func(i, j int) bool { return falseNegatives[i].ID < falseNegatives[j].ID })
	sort.Slice(falsePositives, func(i, j int) bool { return falsePositives[i].ID < falsePositives[j].ID })
	return
}
