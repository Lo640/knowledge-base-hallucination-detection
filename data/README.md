# 数据目录

将测评提供的客服回复放入 `replies.json`，将人工标注放入 `ground_truth.json`。

建议的最小格式：

```json
[
  {"id": "1", "question": "用户问题", "answer": "客服回复"}
]
```

```json
[
  {"id": "1", "hallucination": true, "categories": ["虚构优惠或政策"], "evidence": "知识库证据"}
]
```
