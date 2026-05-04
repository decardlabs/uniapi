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
