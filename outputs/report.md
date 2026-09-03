# 0110 客服回复幻觉检测报告

生成时间：2026-09-03T12:47:07+08:00  
流程：规则初筛 → LLM/人工复核边界 → 风险决策

## 一、项目边界

本项目是离线评测 MVP：输入题目提供的 20 条历史回复、每条对应知识库文本和人工标注，输出逐条检测结果并验证 FN 漏检、FP 误报。`ground_truth.json` 只在评测阶段使用，不参与检测，避免数据泄漏。它不是线上客服拦截服务，也不能用这 20 条样本证明生产泛化能力。

## 二、分类与风险

| 分类 | 严重程度 | 处理建议 |
|---|---|---|
| 安全与合规误导 | P0-严重 | 可能造成健康、安全或合规风险的错误建议。 |
| 能力越界 | P1-高 | 声称完成了系统实际上无法完成的查询、修改或工单操作。 |
| 政策与交易编造 | P1-高 | 虚构或歪曲退货、发票、发货、快递、优惠、支付等政策。 |
| 事实与参数编造 | P2-中 | 产品参数、材质、接口、品牌关系或门店事实与知识库冲突。 |
| 联系方式编造 | P2-中 | 给出知识库不允许或无法核验的地址、电话等信息。 |
| 关键信息遗漏 | P2-中 | 省略知识库重要限制条件，可能导致用户错误决策。 |

P0 直接阻断并转人工；P1 阻断或强制复核；P2 重新检索后改写，无法确认时不要强答。

## 三、验证结果

| 指标 | 数值 | 解释 |
|---|---:|---|
| 样本数 | 20 | 题目要求的 20 条 |
| TP | 18 | 实际有幻觉且成功检出 |
| TN | 2 | 正常回复且未误报 |
| FN / 漏检 | 0 | 实际有幻觉但工具没有检出 |
| FP / 误报 | 0 | 实际正常但工具误报 |
| Precision | 100.00% | 预测为幻觉的结果中有多少准确 |
| Recall | 100.00% | 实际幻觉中有多少被检出 |
| F1 | 100.00% | Precision 与 Recall 的综合值 |

阶段统计：规则确定风险 17 条，送二次复核 1 条，复核闭环 1 条。

## 四、漏检与误报

### 漏检 FN

暂无漏检。

### 误报 FP

暂无误报。

## 五、逐条结果

| ID | 规则状态 | 最终决策 | 预测 | 标注 | 是否正确 | 分类 |
|---|---|---|---|---|---|---|
| h01 | rule_risk | block | true | true | true | 政策与交易编造 |
| h02 | rule_risk | block | true | true | true | 事实与参数编造 |
| h03 | rule_risk | block | true | true | true | 能力越界 |
| h04 | rule_risk | block | true | true | true | 政策与交易编造 |
| h05 | rule_risk | block | true | true | true | 政策与交易编造 |
| h06 | rule_risk | block | true | true | true | 事实与参数编造 |
| h07 | rule_risk | block | true | true | true | 联系方式编造 |
| h08 | rule_risk | block | true | true | true | 政策与交易编造、事实与参数编造 |
| h09 | rule_risk | block | true | true | true | 事实与参数编造 |
| h10 | rule_risk | block | true | true | true | 能力越界 |
| h11 | rule_risk | block | true | true | true | 事实与参数编造 |
| h12 | rule_safe | pass | false | false | true |  |
| h13 | rule_risk | block | true | true | true | 安全与合规误导 |
| h14 | rule_risk | block | true | true | true | 能力越界 |
| h15 | rule_risk | block | true | true | true | 事实与参数编造 |
| h16 | rule_safe | pass | false | false | true |  |
| h17 | rule_risk | block | true | true | true | 事实与参数编造 |
| h18 | rule_risk | block | true | true | true | 能力越界 |
| h19 | rule_risk | block | true | true | true | 政策与交易编造、事实与参数编造 |
| h20 | rule_review | block | true | true | true | 关键信息遗漏 |

## 六、误判原因与边界

本项目即使在当前数据上得到较高分，也不能等同于生产准确率。规则对明确的数值、政策冲突和能力越界可解释且稳定，但会受同义表达、复杂否定、部分正确部分错误、知识库过期和知识库缺失影响。LLM 复核能处理语义和上下文，却也可能产生二次幻觉、受 Prompt 影响、输出不稳定，不能成为唯一裁判。

生产版本还需要：独立测试集和持续人工标注闭环；知识库版本、生效时间和证据 ID；工具调用日志校验；P0/P1/P2 分层策略；线上阻断、转人工、改写和离线复盘；按类别、风险级别、知识库版本监控 FN/FP。

## 七、AI 工具使用情况

使用 Codex 辅助完成工程结构、Go 代码、单元测试、README 和报告模板；人工负责选择方案、定义分类和严重程度、审查 AI 生成代码、确认输入字段、核对人工标注，并复核运行结果。当前 `offline-mock` 只用于演示复核边界，不声称具备真实 LLM 的语义判断能力。
