// Copyright 2026 uniapi authors
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// splitLines 按换行分割字符串
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// containsAll 检查所有子串是否都在 s 中
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

// contains 是子串查找辅助
func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}

// CodexToMiniMaxBasicRequest 模拟 Codex 智能体的基础请求结构
type CodexToMiniMaxBasicRequest struct {
	Model    string     `json:"model"`
	Messages []CodexMsg `json:"messages"`
}

type CodexMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CodexToolUse 表示 Codex 智能体的工具调用结构
type CodexToolUse struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type MiniMaxBasicResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int      `json:"index"`
		Message      CodexMsg `json:"message"`
		FinishReason string   `json:"finish_reason"`
	} `json:"choices"`
}

type MiniMaxFunctionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string        `json:"role"`
			Content string        `json:"content"`
			ToolUse *CodexToolUse `json:"tool_use,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// TestCodexToMiniMax_Streaming 测试流式协议兼容性
func TestCodexToMiniMax_Streaming(t *testing.T) {
	t.Parallel()

	codexReq := CodexToMiniMaxBasicRequest{
		Model:    "codex-1.0",
		Messages: []CodexMsg{{Role: "user", Content: "请分步输出1到3。"}},
	}
	body, err := json.Marshal(codexReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-codex-stream","object":"chat.completion.chunk","created":1714637004,"choices":[{"index":0,"delta":{"role":"assistant","content":"1"}}]}` + "\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-codex-stream","object":"chat.completion.chunk","created":1714637004,"choices":[{"index":0,"delta":{"content":"2"}}]}` + "\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-codex-stream","object":"chat.completion.chunk","created":1714637004,"choices":[{"index":0,"delta":{"content":"3"}}]}` + "\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-codex-stream","object":"chat.completion","created":1714637004,"choices":[{"index":0,"message":{"role":"assistant","content":"1 2 3"},"finish_reason":"stop"}]}` + "\n"))
		w.(http.Flusher).Flush()
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	// 读取流式响应
	var lines []string
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			for _, line := range splitLines(chunk) {
				if len(line) > 0 {
					lines = append(lines, line)
				}
			}
		}
		if err != nil {
			break
		}
	}

	found1 := false
	found2 := false
	found3 := false
	foundFinal := false
	for _, l := range lines {
		if l == "" {
			continue
		}
		if foundFinal {
			continue
		}
		if containsAll(l, []string{"object\":\"chat.completion.chunk", "content\":\"1"}) {
			found1 = true
		}
		if containsAll(l, []string{"object\":\"chat.completion.chunk", "content\":\"2"}) {
			found2 = true
		}
		if containsAll(l, []string{"object\":\"chat.completion.chunk", "content\":\"3"}) {
			found3 = true
		}
		if containsAll(l, []string{"object\":\"chat.completion", "finish_reason\":\"stop"}) {
			foundFinal = true
		}
	}
	require.True(t, found1, "stream chunk 1 not found")
	require.True(t, found2, "stream chunk 2 not found")
	require.True(t, found3, "stream chunk 3 not found")
	require.True(t, foundFinal, "final message not found")
}

// TestCodexToMiniMax_Basic 测试 Codex→MiniMax 基础协议兼容性
func TestCodexToMiniMax_Basic(t *testing.T) {
	t.Parallel()

	codexReq := CodexToMiniMaxBasicRequest{
		Model:    "codex-1.0",
		Messages: []CodexMsg{{Role: "user", Content: "你好，MiniMax！"}},
	}
	body, err := json.Marshal(codexReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-codex-basic","object":"chat.completion","created":1714637000,"choices":[{"index":0,"message":{"role":"assistant","content":"你好，这里是 MiniMax。"},"finish_reason":"stop"}]}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var mmxResp MiniMaxBasicResponse
	require.NoError(t, json.Unmarshal(respBody, &mmxResp))
	require.Equal(t, "chat.completion", mmxResp.Object)
	require.Equal(t, "assistant", mmxResp.Choices[0].Message.Role)
	require.Contains(t, mmxResp.Choices[0].Message.Content, "MiniMax")
}

// TestCodexToMiniMax_FunctionToolUse 测试 Codex function/tool use 映射到 MiniMax
func TestCodexToMiniMax_FunctionToolUse(t *testing.T) {
	t.Parallel()

	codexReq := CodexToMiniMaxBasicRequest{
		Model: "codex-1.0",
		Messages: []CodexMsg{
			{Role: "user", Content: "请调用天气查询工具。"},
		},
	}
	body, err := json.Marshal(codexReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-codex-tool","object":"chat.completion","created":1714637001,"choices":[{"index":0,"message":{"role":"assistant","content":"","tool_use":{"name":"get_weather","arguments":{"city":"北京"}}},"finish_reason":"tool_use"}]}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var mmxResp MiniMaxFunctionResponse
	require.NoError(t, json.Unmarshal(respBody, &mmxResp))
	require.Equal(t, "chat.completion", mmxResp.Object)
	require.Equal(t, "assistant", mmxResp.Choices[0].Message.Role)
	require.NotNil(t, mmxResp.Choices[0].Message.ToolUse)
	require.Equal(t, "get_weather", mmxResp.Choices[0].Message.ToolUse.Name)
	require.Equal(t, "北京", mmxResp.Choices[0].Message.ToolUse.Arguments["city"])
	require.Equal(t, "tool_use", mmxResp.Choices[0].FinishReason)
}

// TestCodexToMiniMax_UsageBillingLogging 测试 Codex→MiniMax 用量与计费日志
func TestCodexToMiniMax_UsageBillingLogging(t *testing.T) {
	t.Parallel()

	codexReq := CodexToMiniMaxBasicRequest{
		Model:    "codex-1.0",
		Messages: []CodexMsg{{Role: "user", Content: "请统计用量和计费。"}},
	}
	body, err := json.Marshal(codexReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-codex-usage","object":"chat.completion","created":1714637003,"choices":[{"index":0,"message":{"role":"assistant","content":"用量和计费信息如下。"},"finish_reason":"stop"}],"usage":{"prompt_tokens":7,"completion_tokens":9,"total_tokens":16},"billing":{"amount":0.002,"currency":"USD"}}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var respMap map[string]interface{}
	require.NoError(t, json.Unmarshal(respBody, &respMap))

	usage, ok := respMap["usage"].(map[string]interface{})
	require.True(t, ok, "usage field missing or wrong type")
	require.EqualValues(t, 7, usage["prompt_tokens"])
	require.EqualValues(t, 9, usage["completion_tokens"])
	require.EqualValues(t, 16, usage["total_tokens"])

	billing, ok := respMap["billing"].(map[string]interface{})
	require.True(t, ok, "billing field missing or wrong type")
	require.EqualValues(t, 0.002, billing["amount"])
	require.Equal(t, "USD", billing["currency"])
	// End of usage/billing logging test
}
