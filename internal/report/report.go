package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/example/knowledge-base-hallucination-detection/internal/detect"
	"github.com/example/knowledge-base-hallucination-detection/internal/model"
)

func Markdown(e model.Evaluation) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# 0110 客服回复幻觉检测报告\n\n生成时间：%s  \n流程：规则初筛 → LLM/人工复核边界 → 风险决策\n\n", e.GeneratedAt)
	b.WriteString("## 一、项目边界\n\n本项目是离线评测 MVP：输入题目提供的 20 条历史回复、每条对应知识库文本和人工标注，输出逐条检测结果并验证 FN 漏检、FP 误报。`ground_truth.json` 只在评测阶段使用，不参与检测，避免数据泄漏。它不是线上客服拦截服务，也不能用这 20 条样本证明生产泛化能力。\n\n")
	b.WriteString("## 二、分类与风险\n\n| 分类 | 严重程度 | 处理建议 |\n|---|---|---|\n")
	for _, category := range detect.Taxonomy {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", category.Name, category.Severity, category.Description)
	}
	b.WriteString("\nP0 直接阻断并转人工；P1 阻断或强制复核；P2 重新检索后改写，无法确认时不要强答。\n\n")
	fmt.Fprintf(&b, "## 三、验证结果\n\n| 指标 | 数值 | 解释 |\n|---|---:|---|\n| 样本数 | %d | 题目要求的 20 条 |\n| TP | %d | 实际有幻觉且成功检出 |\n| TN | %d | 正常回复且未误报 |\n| FN / 漏检 | %d | 实际有幻觉但工具没有检出 |\n| FP / 误报 | %d | 实际正常但工具误报 |\n| Precision | %.2f%% | 预测为幻觉的结果中有多少准确 |\n| Recall | %.2f%% | 实际幻觉中有多少被检出 |\n| F1 | %.2f%% | Precision 与 Recall 的综合值 |\n\n", e.Metrics.Total, e.Metrics.TP, e.Metrics.TN, e.Metrics.FN, e.Metrics.FP, e.Metrics.Precision*100, e.Metrics.Recall*100, e.Metrics.F1*100)
	fmt.Fprintf(&b, "阶段统计：规则确定风险 %d 条，送二次复核 %d 条，复核闭环 %d 条。\n\n", e.StageStats.RuleConfirmed, e.StageStats.SentToReview, e.StageStats.ReviewClosed)
	b.WriteString("## 四、漏检与误报\n\n### 漏检 FN\n\n")
	writeMismatchList(&b, e.FalseNegatives, "暂无漏检。")
	b.WriteString("### 误报 FP\n\n")
	writeMismatchList(&b, e.FalsePositives, "暂无误报。")
	b.WriteString("## 五、逐条结果\n\n| ID | 规则状态 | 最终决策 | 预测 | 标注 | 是否正确 | 分类 |\n|---|---|---|---|---|---|---|\n")
	for _, result := range e.Results {
		fmt.Fprintf(&b, "| %s | %s | %s | %t | %t | %t | %s |\n", result.ID, result.Prediction.Rule.Status, result.Prediction.Decision, result.Prediction.Hallucinate, result.GroundTruth.Hallucinate, result.Correct, strings.Join(result.Prediction.Categories, "、"))
	}
	b.WriteString("\n## 六、误判原因与边界\n\n本项目即使在当前数据上得到较高分，也不能等同于生产准确率。规则对明确的数值、政策冲突和能力越界可解释且稳定，但会受同义表达、复杂否定、部分正确部分错误、知识库过期和知识库缺失影响。LLM 复核能处理语义和上下文，却也可能产生二次幻觉、受 Prompt 影响、输出不稳定，不能成为唯一裁判。\n\n生产版本还需要：独立测试集和持续人工标注闭环；知识库版本、生效时间和证据 ID；工具调用日志校验；P0/P1/P2 分层策略；线上阻断、转人工、改写和离线复盘；按类别、风险级别、知识库版本监控 FN/FP。\n\n")
	b.WriteString("## 七、AI 工具使用情况\n\n使用 Codex 辅助完成工程结构、Go 代码、单元测试、README 和报告模板；人工负责选择方案、定义分类和严重程度、审查 AI 生成代码、确认输入字段、核对人工标注，并复核运行结果。当前 `offline-mock` 只用于演示复核边界，不声称具备真实 LLM 的语义判断能力。\n")
	return b.String()
}

func writeMismatchList(b *strings.Builder, results []model.SampleResult, empty string) {
	if len(results) == 0 {
		b.WriteString(empty + "\n\n")
		return
	}
	for _, result := range results {
		fmt.Fprintf(b, "- `%s`：%s；人工标注=%t，预测=%t；检测原因=%s\n", result.ID, result.Answer, result.GroundTruth.Hallucinate, result.Prediction.Hallucinate, strings.Join(flattenReasons(result.Prediction.Reasons), "；"))
	}
	b.WriteString("\n")
}

func flattenReasons(reasons map[string][]string) []string {
	keys := make([]string, 0, len(reasons))
	for key := range reasons {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out []string
	for _, key := range keys {
		out = append(out, key+"："+strings.Join(reasons[key], "、"))
	}
	return out
}
