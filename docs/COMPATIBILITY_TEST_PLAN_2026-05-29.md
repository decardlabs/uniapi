# DeepSeek/MiniMax 兼容性测试方案（2026-05-29）

## 1. 目标

验证 UniAPI 中转与厂商直连在以下维度的兼容性：

1. 协议兼容：Chat Completions、Responses fallback、SSE 流式。
2. 字段兼容：role、finish_reason、usage、错误结构。
3. 行为兼容：多轮上下文、异常输入处理、超时稳定性。

## 2. 测试对象

1. DeepSeek `deepseek-v4-pro`
2. MiniMax `MiniMax-M2.7`

## 3. 测试路径

1. 直连路径（Direct）：厂商 BaseURL + 厂商 API Key。
2. 中转路径（Relay）：`http://localhost:3000` + UniAPI Token。

## 4. 用例矩阵

每个模型执行以下用例：

1. 非流式基础问答。
2. 流式输出（SSE chunk + DONE）。
3. 多轮上下文连续对话。
4. Responses API（或 fallback）路径。
5. 异常输入（非法 role）错误映射。

## 5. 对比指标

1. 成功率：HTTP 2xx 占比。
2. 延迟：首包与总耗时（手工记录）。
3. 结构一致性：关键字段是否齐全。
4. 行为一致性：响应语义是否可用。
5. 错误一致性：错误码、错误信息可诊断性。

## 6. 验收口径（软完整）

1. 中转成功率不低于直连 98%。
2. 关键字段缺失率 <= 1%。
3. 中转延迟 P95 不高于直连 1.25 倍（样本量允许时）。
4. MiniMax 不出现 `message[i].role must be 'user' or 'assistant'`。

## 7. 输出

测试执行记录与结论写入：

1. `docs/COMPATIBILITY_TEST_RESULTS_2026-05-29.md`
