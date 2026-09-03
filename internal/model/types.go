package model

type Reply struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	Answer        string `json:"answer"`
	KnowledgeBase string `json:"knowledge_base,omitempty"`
}

type GroundTruth struct {
	ID          string   `json:"id"`
	Hallucinate bool     `json:"hallucination"`
	Categories  []string `json:"categories,omitempty"`
	Evidence    string   `json:"evidence,omitempty"`
}

type RuleResult struct {
	Status      string              `json:"status"`
	Categories  []string            `json:"categories,omitempty"`
	Reasons     map[string][]string `json:"reasons,omitempty"`
	ReviewHints []string            `json:"review_hints,omitempty"`
}

type LLMReview struct {
	Status      string   `json:"status"`
	Verdict     string   `json:"verdict"`
	Hallucinate bool     `json:"hallucination"`
	Categories  []string `json:"categories,omitempty"`
	Reason      string   `json:"reason,omitempty"`
	Provider    string   `json:"provider"`
}

type Detection struct {
	Hallucinate bool                `json:"hallucination"`
	Verdict     string              `json:"verdict"`
	Categories  []string            `json:"categories,omitempty"`
	Reasons     map[string][]string `json:"reasons,omitempty"`
	Confidence  float64             `json:"confidence"`
	Severity    string              `json:"severity,omitempty"`
	Decision    string              `json:"decision"`
	Stage       string              `json:"stage"`
	Rule        RuleResult          `json:"rule_result"`
	LLM         *LLMReview          `json:"llm_review,omitempty"`
}

type SampleResult struct {
	ID          string      `json:"id"`
	Question    string      `json:"question"`
	Answer      string      `json:"answer"`
	Prediction  Detection   `json:"prediction"`
	GroundTruth GroundTruth `json:"ground_truth"`
	Correct     bool        `json:"correct"`
}

type Metrics struct {
	Total     int     `json:"total"`
	TP        int     `json:"true_positive"`
	TN        int     `json:"true_negative"`
	FP        int     `json:"false_positive"`
	FN        int     `json:"false_negative"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1        float64 `json:"f1"`
}

type StageStats struct {
	RuleConfirmed int `json:"rule_confirmed"`
	SentToReview  int `json:"sent_to_review"`
	ReviewClosed  int `json:"review_closed"`
}

type Evaluation struct {
	GeneratedAt    string         `json:"generated_at"`
	Mode           string         `json:"mode"`
	Metrics        Metrics        `json:"metrics"`
	StageStats     StageStats     `json:"stage_stats"`
	ByCategory     map[string]int `json:"predicted_by_category"`
	Results        []SampleResult `json:"results"`
	FalseNegatives []SampleResult `json:"false_negatives"`
	FalsePositives []SampleResult `json:"false_positives"`
}
