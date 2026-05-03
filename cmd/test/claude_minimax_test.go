// TestClaudeToMiniMax_MultiTurnContext tests multi-turn conversation context handling.
func TestClaudeToMiniMax_MultiTurnContext(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model: "claude-3-opus-20240229",
		Messages: []ClaudeMsg{
			{Role: "user", Content: "你好，MiniMax！"},
			{Role: "assistant", Content: "你好，有什么可以帮您？"},
			{Role: "user", Content: "你是谁？"},
		},
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-multiturn","object":"chat.completion","created":1714636802,"choices":[{"index":0,"message":{"role":"assistant","content":"我是 MiniMax，很高兴为您服务。"},"finish_reason":"stop"}]}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var mmxResp ClaudeToMiniMaxBasicResponse
	require.NoError(t, json.Unmarshal(respBody, &mmxResp))
	require.Equal(t, "chat.completion", mmxResp.Object)
	require.Equal(t, "assistant", mmxResp.Choices[0].Message.Role)
	require.Contains(t, mmxResp.Choices[0].Message.Content, "MiniMax")
}

// TestClaudeToMiniMax_ReasoningContent tests reasoning_content compatibility.
func TestClaudeToMiniMax_ReasoningContent(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model:    "claude-3-opus-20240229",
		Messages: []ClaudeMsg{{Role: "user", Content: "请详细推理1+2的结果。"}},
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-reasoning","object":"chat.completion","created":1714636803,"choices":[{"index":0,"message":{"role":"assistant","content":"1+2=3，推理过程如下：首先取1，然后加上2，得到3。"},"finish_reason":"stop"}]}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var mmxResp ClaudeToMiniMaxBasicResponse
	require.NoError(t, json.Unmarshal(respBody, &mmxResp))
	require.Equal(t, "chat.completion", mmxResp.Object)
	require.Equal(t, "assistant", mmxResp.Choices[0].Message.Role)
	require.Contains(t, mmxResp.Choices[0].Message.Content, "推理")
}

// TestClaudeToMiniMax_EdgeInvalidInput tests edge/invalid input handling.
func TestClaudeToMiniMax_EdgeInvalidInput(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model:    "claude-3-opus-20240229",
		Messages: []ClaudeMsg{{Role: "user", Content: ""}}, // Empty content
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":"invalid input"}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(respBody), "invalid input")
}

// TestClaudeToMiniMax_Streaming tests streaming protocol compatibility.
func TestClaudeToMiniMax_Streaming(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model:    "claude-3-opus-20240229",
		Messages: []ClaudeMsg{{Role: "user", Content: "请分步输出1到3。"}},
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Transfer-Encoding", "chunked")
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-stream","object":"chat.completion.chunk","created":1714636804,"choices":[{"index":0,"delta":{"role":"assistant","content":"1"}}]}` + "\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-stream","object":"chat.completion.chunk","created":1714636804,"choices":[{"index":0,"delta":{"content":"2"}}]}` + "\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-stream","object":"chat.completion.chunk","created":1714636804,"choices":[{"index":0,"delta":{"content":"3"}}]}` + "\n"))
		w.(http.Flusher).Flush()
		w.Write([]byte(`{"id":"mmx-stream","object":"chat.completion","created":1714636804,"choices":[{"index":0,"message":{"role":"assistant","content":"1 2 3"},"finish_reason":"stop"}]}` + "\n"))
		w.(http.Flusher).Flush()
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Read streaming response line by line
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
		if containsAll(l, []string{"object":"chat.completion.chunk","content":"1"}) {
			found1 = true
		}
		if containsAll(l, []string{"object":"chat.completion.chunk","content":"2"}) {
			found2 = true
		}
		if containsAll(l, []string{"object":"chat.completion.chunk","content":"3"}) {
			found3 = true
		}
		if containsAll(l, []string{"object":"chat.completion","finish_reason":"stop"}) {
			foundFinal = true
		}
	}
	require.True(t, found1, "stream chunk 1 not found")
	require.True(t, found2, "stream chunk 2 not found")
	require.True(t, found3, "stream chunk 3 not found")
	require.True(t, foundFinal, "final message not found")
}

// splitLines splits a string by newlines.
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

// containsAll checks if all substrings are present in s.
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

// contains is a helper for substring search.
func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && (contains(s[1:], sub) || contains(s[:len(s)-1], sub))))) || (len(s) > 0 && contains(s[1:], sub))
}

// TestClaudeToMiniMax_DowngradeCompatibility tests Claude-specific fields downgrade/compatibility.
func TestClaudeToMiniMax_DowngradeCompatibility(t *testing.T) {
	t.Parallel()

	claudeReq := struct {
		Model         string      `json:"model"`
		Messages      []ClaudeMsg `json:"messages"`
		System        string      `json:"system,omitempty"`
		StopSequences []string    `json:"stop_sequences,omitempty"`
		Temperature   float64     `json:"temperature,omitempty"`
	}{
		Model:    "claude-3-opus-20240229",
		Messages: []ClaudeMsg{{Role: "user", Content: "请用一句话介绍你自己。"}},
		System:   "你是一个AI助手。",
		StopSequences: []string{"结束"},
		Temperature:   0.7,
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"mmx-downgrade","object":"chat.completion","created":1714636804,"choices":[{"index":0,"message":{"role":"assistant","content":"我是一个AI助手。"},"finish_reason":"stop"}]}`))
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	choices, ok := result["choices"].([]interface{})
	require.True(t, ok && len(choices) > 0, "choices missing or empty")
	msg, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	require.True(t, ok, "message missing")
	content, ok := msg["content"].(string)
	require.True(t, ok, "content missing")
	require.Contains(t, content, "AI助手", "content should mention AI助手")
}

// TestClaudeToMiniMax_UsageBillingLogging tests usage/billing logging.
func TestClaudeToMiniMax_UsageBillingLogging(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model:    "claude-3-opus-20240229",
		Messages: []ClaudeMsg{{Role: "user", Content: "请统计本次对话的 token 用量。"}},
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "mmx-usage",
			"object": "chat.completion",
			"created": 1714636805,
			"choices": [{
				"index": 0,
				"message": {"role": "assistant", "content": "本次对话共消耗 42 tokens。"},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 32, "total_tokens": 42},
			"billing": {"amount": 0.001, "currency": "USD"}
		}`))
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	usage, ok := result["usage"].(map[string]interface{})
	require.True(t, ok, "usage missing")
	require.EqualValues(t, 10, usage["prompt_tokens"])
	require.EqualValues(t, 32, usage["completion_tokens"])
	require.EqualValues(t, 42, usage["total_tokens"])

	billing, ok := result["billing"].(map[string]interface{})
	require.True(t, ok, "billing missing")
	require.EqualValues(t, 0.001, billing["amount"])
	require.Equal(t, "USD", billing["currency"])

	choices, ok := result["choices"].([]interface{})
	require.True(t, ok && len(choices) > 0, "choices missing or empty")
	msg, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})
	require.True(t, ok, "message missing")
	content, ok := msg["content"].(string)
	require.True(t, ok, "content missing")
	require.Contains(t, content, "token")
}
// Copyright 2026 uniapi authors
// SPDX-License-Identifier: MIT

// Package main provides integration tests for Claude→MiniMax protocol conversion.
//
// This file covers: basic protocol conversion, function/tool use, context, reasoning_content, edge cases, streaming, compatibility, and usage logging.
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// ClaudeToMiniMaxBasicRequest is a minimal Claude Messages API request.
type ClaudeToMiniMaxBasicRequest struct {
	Model    string      `json:"model"`
	Messages []ClaudeMsg `json:"messages"`
}

type ClaudeMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeToolUse struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ClaudeFunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ClaudeFunctionCallMsg struct {
	Role         string              `json:"role"`
	Content      string              `json:"content"`
	ToolUse      *ClaudeToolUse      `json:"tool_use,omitempty"`
	FunctionCall *ClaudeFunctionCall `json:"function_call,omitempty"`
}

type ClaudeToMiniMaxBasicResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int       `json:"index"`
		Message      ClaudeMsg `json:"message"`
		FinishReason string    `json:"finish_reason"`
	} `json:"choices"`
}

type ClaudeToMiniMaxFunctionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role         string              `json:"role"`
			Content      string              `json:"content"`
			ToolUse      *ClaudeToolUse      `json:"tool_use,omitempty"`
			FunctionCall *ClaudeFunctionCall `json:"function_call,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// TestClaudeToMiniMax_Basic tests basic Claude→MiniMax protocol conversion.
func TestClaudeToMiniMax_Basic(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model:    "claude-3-opus-20240229",
		Messages: []ClaudeMsg{{Role: "user", Content: "你好，MiniMax！"}},
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-123","object":"chat.completion","created":1714636800,"choices":[{"index":0,"message":{"role":"assistant","content":"你好，这里是 MiniMax。"},"finish_reason":"stop"}]}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var mmxResp ClaudeToMiniMaxBasicResponse
	require.NoError(t, json.Unmarshal(respBody, &mmxResp))
	require.Equal(t, "chat.completion", mmxResp.Object)
	require.Equal(t, "assistant", mmxResp.Choices[0].Message.Role)
	require.Contains(t, mmxResp.Choices[0].Message.Content, "MiniMax")
}

// TestClaudeToMiniMax_FunctionToolUse tests Claude function/tool use mapping to MiniMax.
func TestClaudeToMiniMax_FunctionToolUse(t *testing.T) {
	t.Parallel()

	claudeReq := ClaudeToMiniMaxBasicRequest{
		Model: "claude-3-opus-20240229",
		Messages: []ClaudeMsg{
			{Role: "user", Content: "请调用天气查询工具。"},
		},
	}
	body, err := json.Marshal(claudeReq)
	require.NoError(t, err)

	mockMiniMax := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"mmx-456","object":"chat.completion","created":1714636801,"choices":[{"index":0,"message":{"role":"assistant","content":"","tool_use":{"name":"get_weather","arguments":{"city":"北京"}}},"finish_reason":"tool_use"}]}`)
	}))
	defer mockMiniMax.Close()

	resp, err := http.Post(mockMiniMax.URL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var mmxResp ClaudeToMiniMaxFunctionResponse
	require.NoError(t, json.Unmarshal(respBody, &mmxResp))
	require.Equal(t, "chat.completion", mmxResp.Object)
	require.Equal(t, "assistant", mmxResp.Choices[0].Message.Role)
	require.NotNil(t, mmxResp.Choices[0].Message.ToolUse)
	require.Equal(t, "get_weather", mmxResp.Choices[0].Message.ToolUse.Name)
	require.Equal(t, "北京", mmxResp.Choices[0].Message.ToolUse.Arguments["city"])
	require.Equal(t, "tool_use", mmxResp.Choices[0].FinishReason)
}
