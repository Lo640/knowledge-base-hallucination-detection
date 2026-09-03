package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/knowledge-base-hallucination-detection/internal/detect"
	"github.com/example/knowledge-base-hallucination-detection/internal/eval"
	"github.com/example/knowledge-base-hallucination-detection/internal/loader"
	"github.com/example/knowledge-base-hallucination-detection/internal/model"
	"github.com/example/knowledge-base-hallucination-detection/internal/report"
)

func main() {
	input := flag.String("input", "data/replies.json", "客服回复 JSON 文件")
	truthPath := flag.String("ground-truth", "data/ground_truth.json", "人工标注 JSON 文件")
	outputDir := flag.String("output-dir", "outputs", "输出目录")
	mode := flag.String("mode", "mock", "复核模式：mock 或 rule-only")
	flag.Parse()
	if *mode != "mock" && *mode != "rule-only" {
		fail(fmt.Errorf("unsupported mode %q", *mode))
	}

	replies, err := loader.LoadReplies(*input)
	failIf(err)
	truths, err := loader.LoadGroundTruth(*truthPath)
	failIf(err)
	truthByID := make(map[string]model.GroundTruth, len(truths))
	for _, truth := range truths {
		truthByID[truth.ID] = truth
	}

	var reviewer detect.ReviewModel
	if *mode == "mock" {
		reviewer = detect.OfflineReviewModel{}
	}
	pipeline := detect.Pipeline{Rules: detect.RuleDetector{}, Reviewer: reviewer}
	results := make([]model.SampleResult, 0, len(replies))
	byCategory := make(map[string]int)
	stats := model.StageStats{}
	for _, reply := range replies {
		prediction := pipeline.Detect(reply)
		truth, exists := truthByID[reply.ID]
		if !exists {
			truth = model.GroundTruth{ID: reply.ID}
		}
		if prediction.Rule.Status == detect.RuleRisk {
			stats.RuleConfirmed++
		}
		if prediction.Rule.Status == detect.RuleReview {
			stats.SentToReview++
		}
		if prediction.LLM != nil {
			stats.ReviewClosed++
		}
		for _, category := range prediction.Categories {
			byCategory[category]++
		}
		results = append(results, model.SampleResult{
			ID: reply.ID, Question: reply.Question, Answer: reply.Answer,
			Prediction: prediction, GroundTruth: truth,
			Correct: prediction.Hallucinate == truth.Hallucinate,
		})
	}
	fn, fp := eval.SplitMismatches(results)
	evaluation := model.Evaluation{
		GeneratedAt: time.Now().Format(time.RFC3339), Mode: *mode,
		Metrics: eval.Calculate(results), StageStats: stats,
		ByCategory: byCategory, Results: results,
		FalseNegatives: fn, FalsePositives: fp,
	}
	failIf(os.MkdirAll(*outputDir, 0755))
	failIf(writeJSON(filepath.Join(*outputDir, "result.json"), evaluation))
	failIf(os.WriteFile(filepath.Join(*outputDir, "report.md"), []byte(report.Markdown(evaluation)), 0644))
	fmt.Printf("完成：%d 条样本，TP=%d TN=%d FN=%d FP=%d Recall=%.2f%% Precision=%.2f%% F1=%.2f%%\n", evaluation.Metrics.Total, evaluation.Metrics.TP, evaluation.Metrics.TN, evaluation.Metrics.FN, evaluation.Metrics.FP, evaluation.Metrics.Recall*100, evaluation.Metrics.Precision*100, evaluation.Metrics.F1*100)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

func failIf(err error) {
	if err != nil {
		fail(err)
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, "错误:", err); os.Exit(1) }
