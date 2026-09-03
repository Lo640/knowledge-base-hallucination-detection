package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/example/knowledge-base-hallucination-detection/internal/model"
)

func LoadReplies(path string) ([]model.Reply, error) {
	records, err := readRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]model.Reply, 0, len(records))
	for i, record := range records {
		id := firstString(record, "id", "uuid", "sample_id", "qid", "question_id")
		if id == "" {
			id = fmt.Sprintf("sample-%03d", i+1)
		}
		out = append(out, model.Reply{
			ID:            id,
			Question:      firstString(record, "question", "query", "user_question", "prompt"),
			Answer:        firstString(record, "answer", "reply", "response", "system_reply", "content", "text"),
			KnowledgeBase: firstString(record, "knowledge_base", "knowledgeBase", "kb", "context"),
		})
	}
	return out, nil
}

func LoadGroundTruth(path string) ([]model.GroundTruth, error) {
	records, err := readRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]model.GroundTruth, 0, len(records))
	for i, record := range records {
		id := firstString(record, "id", "uuid", "sample_id", "qid", "question_id")
		if id == "" {
			id = fmt.Sprintf("sample-%03d", i+1)
		}
		out = append(out, model.GroundTruth{
			ID:          id,
			Hallucinate: firstBool(record, "hallucination", "is_hallucination", "isHallucination", "label", "is_hallucinated"),
			Categories:  firstStrings(record, "categories", "category", "hallucination_type", "type"),
			Evidence:    firstString(record, "evidence", "reason", "explanation", "gold_reason", "detail"),
		})
	}
	return out, nil
}

func readRecords(path string) ([]map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if list, ok := raw.([]any); ok {
		return maps(list)
	}
	if obj, ok := raw.(map[string]any); ok {
		for _, key := range []string{"data", "items", "records", "replies", "ground_truth", "groundTruth"} {
			if list, ok := obj[key].([]any); ok {
				return maps(list)
			}
		}
	}
	return nil, fmt.Errorf("%s must contain a JSON array or a wrapper with data/items/records", path)
}

func maps(list []any) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(list))
	for i, item := range list {
		obj, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("record %d is not a JSON object", i+1)
		}
		out = append(out, obj)
	}
	return out, nil
}

func firstString(record map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := record[key]; ok {
			if text, ok := value.(string); ok {
				return strings.TrimSpace(text)
			}
			if value != nil {
				return fmt.Sprint(value)
			}
		}
	}
	return ""
}

func firstStrings(record map[string]any, keys ...string) []string {
	for _, key := range keys {
		value, ok := record[key]
		if !ok {
			continue
		}
		switch values := value.(type) {
		case []any:
			out := make([]string, 0, len(values))
			for _, value := range values {
				out = append(out, fmt.Sprint(value))
			}
			return out
		case string:
			if strings.TrimSpace(values) == "" {
				return nil
			}
			return []string{strings.TrimSpace(values)}
		}
	}
	return nil
}

func firstBool(record map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := record[key]
		if !ok {
			continue
		}
		switch value := value.(type) {
		case bool:
			return value
		case float64:
			return value != 0
		case string:
			switch strings.ToLower(strings.TrimSpace(value)) {
			case "true", "1", "yes", "y", "有", "是", "幻觉", "hallucination":
				return true
			case "false", "0", "no", "n", "无", "否", "正常", "non-hallucination":
				return false
			}
		}
	}
	return false
}
