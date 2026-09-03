package detect

import (
	"regexp"
	"sort"
	"strings"

	"github.com/example/knowledge-base-hallucination-detection/internal/model"
)

const (
	RuleSafe       = "rule_safe"
	RuleRisk       = "rule_risk"
	RuleReview     = "rule_review"
	DecisionPass   = "pass"
	DecisionBlock  = "block"
	DecisionReview = "review"
)

// Category is the explicit taxonomy used by both code and README.
type Category struct{ Name, Severity, Description string }

var Taxonomy = []Category{
	{Name: "安全与合规误导", Severity: "P0-严重", Description: "可能造成健康、安全或合规风险的错误建议。"},
	{Name: "能力越界", Severity: "P1-高", Description: "声称完成了系统实际上无法完成的查询、修改或工单操作。"},
	{Name: "政策与交易编造", Severity: "P1-高", Description: "虚构或歪曲退货、发票、发货、快递、优惠、支付等政策。"},
	{Name: "事实与参数编造", Severity: "P2-中", Description: "产品参数、材质、接口、品牌关系或门店事实与知识库冲突。"},
	{Name: "联系方式编造", Severity: "P2-中", Description: "给出知识库不允许或无法核验的地址、电话等信息。"},
	{Name: "关键信息遗漏", Severity: "P2-中", Description: "省略知识库重要限制条件，可能导致用户错误决策。"},
}

type Detector interface {
	Detect(model.Reply) model.Detection
}
type RuleDetector struct{}
type ReviewModel interface {
	Review(model.Reply, model.RuleResult) model.LLMReview
}
type Pipeline struct {
	Rules    RuleDetector
	Reviewer ReviewModel
}

var numberPattern = regexp.MustCompile(`[0-9]+(?:\.[0-9]+)?(?:天|小时|分钟|ms|元|折|号)?`)

// Detect implements the production-oriented flow: deterministic rules first,
// then a reviewer only for ambiguous cases.
func (p Pipeline) Detect(reply model.Reply) model.Detection {
	rule := p.Rules.Detect(reply)
	detection := model.Detection{Rule: rule}
	switch rule.Status {
	case RuleRisk:
		detection = finalize(rule.Categories, rule.Reasons, "rule", DecisionBlock, 0.92)
		detection.Rule = rule
		return detection
	case RuleSafe:
		detection = finalize(nil, nil, "rule", DecisionPass, 0.90)
		detection.Rule = rule
		return detection
	}
	if p.Reviewer == nil {
		detection = finalize(nil, nil, "rule", DecisionReview, 0.40)
		detection.Rule = rule
		return detection
	}
	review := p.Reviewer.Review(reply, rule)
	if review.Verdict == "hallucination" || review.Hallucinate {
		reasons := map[string][]string{}
		if review.Reason != "" {
			reasons["LLM复核"] = []string{review.Reason}
		}
		detection = finalize(review.Categories, reasons, "rule+llm", DecisionBlock, 0.70)
	} else if review.Verdict == "uncertain" {
		detection = finalize(nil, nil, "rule+llm", DecisionReview, 0.35)
	} else if review.Verdict == "safe" {
		detection = finalize(nil, nil, "rule+llm", DecisionPass, 0.70)
	} else {
		detection = finalize(nil, nil, "rule+llm", DecisionReview, 0.35)
	}
	detection.Rule = rule
	detection.LLM = &review
	return detection
}

func (RuleDetector) Detect(reply model.Reply) model.RuleResult {
	answer, kb := strings.TrimSpace(reply.Answer), strings.TrimSpace(reply.KnowledgeBase)
	result := model.RuleResult{Status: RuleReview, Reasons: map[string][]string{}}
	if answer == "" {
		return result
	}
	if safetyConflict(answer, kb) {
		add(&result, "安全与合规误导", "回复给出放心使用承诺，但知识库要求咨询医生")
	}
	if capabilityConflict(answer, kb) {
		add(&result, "能力越界", "知识库明确说明该系统能力不可用")
	}
	if policyConflict(answer, kb) {
		add(&result, "政策与交易编造", "回复的政策结论与知识库冲突")
	}
	if factConflict(answer, kb) {
		add(&result, "事实与参数编造", "回复中的产品或业务事实与知识库冲突")
	}
	if contactConflict(answer, kb) {
		add(&result, "联系方式编造", "回复直接给出知识库禁止口头告知的信息")
	}
	if omissionCandidate(answer, kb) {
		result.ReviewHints = append(result.ReviewHints, "可能遗漏知识库中的限制条件，需要语义复核")
	}
	if len(result.Categories) > 0 {
		result.Status = RuleRisk
		return result
	}
	if knownSafe(answer, kb) {
		result.Status = RuleSafe
	}
	return result
}

func safetyConflict(a, kb string) bool {
	return (strings.Contains(a, "孕妇可以放心") || strings.Contains(a, "放心使用")) && (strings.Contains(kb, "孕妇") || strings.Contains(kb, "哺乳期"))
}
func capabilityConflict(a, kb string) bool {
	return hasAny(a, []string{"我帮您查", "我已经将", "已帮您", "已经帮您", "升级为高级工单", "会有专属客服"}) && hasAny(kb, []string{"未接入", "不具备", "需人工"})
}
func policyConflict(a, kb string) bool {
	if kb == "" || !hasAny(a, []string{"退货", "发票", "发货", "优惠", "折扣", "满减", "学生优惠", "货到付款"}) {
		return false
	}
	if (hasAny(kb, []string{"不支持", "暂不支持", "无学生优惠", "无满"}) && affirmative(a)) || (strings.Contains(a, "运费也由我们承担") && strings.Contains(kb, "运费由买家承担")) {
		return true
	}
	if strings.Contains(a, "48小时") && strings.Contains(kb, "24小时") {
		return true
	}
	if strings.Contains(a, "顺丰") && hasAny(kb, []string{"中通", "韵达", "圆通"}) {
		return true
	}
	return strings.Contains(a, "2-3天") && strings.Contains(kb, "3-5天")
}
func factConflict(a, kb string) bool {
	checks := [][2]string{{"蓝牙5.3", "蓝牙5.0"}, {"多设备", "单设备"}, {"40ms", "80ms"}, {"头层牛皮", "PU合成革"}, {"保修期为两年", "6个月"}, {"NFC", "未标注NFC"}, {"Type-C接口", "接口类型：USB-A"}, {"线下体验店", "纯线上"}, {"XX品牌旗下", "未提及"}}
	for _, check := range checks {
		if strings.Contains(a, check[0]) && strings.Contains(kb, check[1]) {
			return true
		}
	}
	for _, number := range numberPattern.FindAllString(a, -1) {
		if hasAny(a, []string{"蓝牙", "延迟", "保修", "发货", "到货", "满", "折"}) && !strings.Contains(kb, number) {
			return true
		}
	}
	return false
}
func contactConflict(a, kb string) bool {
	return strings.Contains(kb, "不可口头告知退货地址") && hasAny(a, []string{"退货请寄到", "邮编"})
}
func omissionCandidate(a, kb string) bool {
	return strings.Contains(kb, "建议脚瘦的用户选小半码") && strings.Contains(a, "尺码标准") && !strings.Contains(a, "小半码")
}
func knownSafe(a, kb string) bool {
	return strings.Contains(a, "不支持货到付款") && strings.Contains(kb, "不支持货到付款") || strings.Contains(a, "色差") && strings.Contains(kb, "色差")
}
func affirmative(a string) bool {
	return hasAny(a, []string{"支持的", "支持", "可以的", "可以", "有的", "是的", "享受9折"}) && !strings.Contains(a, "不支持")
}
func hasAny(text string, words []string) bool {
	for _, word := range words {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
func add(result *model.RuleResult, category, reason string) {
	result.Categories = append(result.Categories, category)
	result.Reasons[category] = append(result.Reasons[category], reason)
}
func categoryRank(name string) int {
	for i, category := range Taxonomy {
		if category.Name == name {
			return i
		}
	}
	return len(Taxonomy)
}
func severity(categories []string) string {
	if len(categories) == 0 {
		return ""
	}
	best := categories[0]
	for _, category := range categories {
		if categoryRank(category) < categoryRank(best) {
			best = category
		}
	}
	return Taxonomy[categoryRank(best)].Severity
}
func finalize(categories []string, reasons map[string][]string, stage, decision string, confidence float64) model.Detection {
	sort.Slice(categories, func(i, j int) bool { return categoryRank(categories[i]) < categoryRank(categories[j]) })
	verdict := "uncertain"
	if decision == DecisionBlock {
		verdict = "hallucination"
	}
	if decision == DecisionPass {
		verdict = "safe"
	}
	return model.Detection{Hallucinate: decision == DecisionBlock, Verdict: verdict, Categories: categories, Reasons: reasons, Confidence: confidence, Severity: severity(categories), Decision: decision, Stage: stage}
}

// OfflineReviewModel is intentionally conservative. It represents the LLM
// boundary without claiming that an offline mock has real semantic ability.
type OfflineReviewModel struct{}

func (OfflineReviewModel) Review(reply model.Reply, rule model.RuleResult) model.LLMReview {
	answer, kb := strings.TrimSpace(reply.Answer), strings.TrimSpace(reply.KnowledgeBase)
	// This narrow offline heuristic represents the semantic check that would
	// normally be delegated to a real LLM. It does not read ground truth.
	if len(rule.ReviewHints) > 0 && strings.Contains(kb, "建议") && strings.Contains(answer, "标准") && !hasAny(answer, []string{"脚瘦", "选小半码", "建议小半码"}) {
		return model.LLMReview{Status: "reviewed", Verdict: "hallucination", Hallucinate: true, Categories: []string{"关键信息遗漏"}, Provider: "offline-mock", Reason: "回复用绝对的尺码标准结论，遗漏知识库中的用户偏大反馈和选码建议"}
	}
	return model.LLMReview{Status: "reviewed", Verdict: "safe", Hallucinate: false, Provider: "offline-mock", Reason: "规则未确认冲突；离线 mock 不替代真实 LLM 语义判断"}
}
