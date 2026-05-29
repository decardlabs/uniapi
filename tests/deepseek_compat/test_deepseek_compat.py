"""
DeepSeek Compatibility Test Suite
===================================
Tests Claude Messages API requests routed through UniAPI → DeepSeek
against direct DeepSeek API calls.

Usage:
    export UNIAPI_BASE_URL="http://localhost:3000"
    export UNIAPI_API_KEY="sk-your-uniapi-key"
    export DEEPSEEK_API_KEY="sk-your-deepseek-key"

    python test_deepseek_compat.py
    python test_deepseek_compat.py --model deepseek-v4-flash
    python test_deepseek_compat.py --verbose
"""

import json
import math
import os
import sys
import time
import argparse
import traceback
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Any, Optional

import requests

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

UNIAPI_BASE_URL = os.environ.get("UNIAPI_BASE_URL", "http://localhost:3000")
UNIAPI_API_KEY = os.environ.get("UNIAPI_API_KEY", "sk-test-key")
DEEPSEEK_API_KEY = os.environ.get("DEEPSEEK_API_KEY", "sk-your-deepseek-key")
DEEPSEEK_BASE_URL = "https://api.deepseek.com"

MODELS = ["deepseek-v4-pro", "deepseek-v4-flash"]

# Prompt tokens should be identical (same input to same model).
# Allow 10% as safety margin for tokenizer edge cases.
PROMPT_TOKEN_TOLERANCE = 0.10

# Completion tokens vary between calls because the model generates
# different text each time (even with temperature=0, there's some variance).
# We use a wider tolerance and mainly verify prompt_tokens and non-zero.
COMPLETION_TOKEN_TOLERANCE = 0.50


# ---------------------------------------------------------------------------
# Stop reason mapping: Claude ↔ OpenAI
# ---------------------------------------------------------------------------

# Claude stop_reason → OpenAI finish_reason equivalence
STOP_REASON_MAP = {
    "end_turn":      "stop",
    "tool_use":      "tool_calls",
    "max_tokens":    "length",
    "stop_sequence": "stop",
}


def stop_reasons_match(claude_stop: Optional[str], openai_stop: Optional[str]) -> bool:
    """Check if Claude stop_reason is semantically equivalent to OpenAI finish_reason."""
    if claude_stop is None and openai_stop is None:
        return True
    if claude_stop is None or openai_stop is None:
        return False

    # Direct match
    if claude_stop == openai_stop:
        return True

    # Map Claude → OpenAI equivalent and compare
    mapped = STOP_REASON_MAP.get(claude_stop, claude_stop)
    return mapped == openai_stop


# ---------------------------------------------------------------------------
# Data Classes
# ---------------------------------------------------------------------------

@dataclass
class TestResult:
    name: str
    category: str
    model: str
    passed: bool
    duration_ms: float
    uniapi_status: Optional[int] = None
    deepseek_status: Optional[int] = None
    uniapi_tokens: Optional[dict] = None
    deepseek_tokens: Optional[dict] = None
    billing_check: Optional[str] = None  # "pass", "fail", "skip"
    error: Optional[str] = None
    details: dict = field(default_factory=dict)


# ---------------------------------------------------------------------------
# Request Builders
# ---------------------------------------------------------------------------

def make_claude_request(model: str, **kwargs) -> dict:
    """Build a Claude Messages API request body."""
    req = {
        "model": model,
        "max_tokens": kwargs.pop("max_tokens", 1024),
        "messages": kwargs.pop("messages", [{"role": "user", "content": "Hello"}]),
    }
    for key in ("system", "tools", "tool_choice", "thinking",
                "temperature", "top_p", "stream", "stop_sequences"):
        if key in kwargs:
            val = kwargs.pop(key)
            if val is not None:
                req[key] = val
    req.update(kwargs)
    return req


def make_openai_request(model: str, **kwargs) -> dict:
    """Build an OpenAI Chat Completions request."""
    req = {
        "model": model,
        "max_tokens": kwargs.pop("max_tokens", 1024),
        "messages": kwargs.pop("messages", [{"role": "user", "content": "Hello"}]),
    }
    for key in ("tools", "tool_choice", "temperature", "top_p", "stream"):
        if key in kwargs:
            val = kwargs.pop(key)
            if val is not None:
                req[key] = val
    if "stop" in kwargs:
        req["stop"] = kwargs.pop("stop")
    req.update(kwargs)
    return req


# ---------------------------------------------------------------------------
# API Clients
# ---------------------------------------------------------------------------

class UniAPIClient:
    """Client for UniAPI Claude Messages endpoint."""

    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.session = requests.Session()

    def _headers(self) -> dict:
        return {
            "x-api-key": self.api_key,
            "anthropic-version": "2023-06-01",
            "content-type": "application/json",
            "accept": "application/json",
        }

    def chat(self, body: dict, timeout: int = 90) -> requests.Response:
        url = f"{self.base_url}/v1/messages"
        return self.session.post(
            url, headers=self._headers(), json=body, timeout=timeout
        )

    def chat_stream(self, body: dict, timeout: int = 120) -> list[dict]:
        body = {**body, "stream": True}
        url = f"{self.base_url}/v1/messages"
        response = self.session.post(
            url, headers=self._headers(), json=body, timeout=timeout, stream=True
        )
        response.raise_for_status()
        events = []
        for line in response.iter_lines(decode_unicode=True):
            if not line:
                continue
            if line.startswith("data: "):
                data = line[6:]
                if data.strip() == "[DONE]":
                    break
                try:
                    events.append(json.loads(data))
                except json.JSONDecodeError:
                    pass
            elif line.startswith("event: "):
                events.append({"__event__": line[7:]})
        return events


class DeepSeekClient:
    """Client for direct DeepSeek API calls (OpenAI format)."""

    def __init__(self, api_key: str):
        self.base_url = DEEPSEEK_BASE_URL
        self.api_key = api_key
        self.session = requests.Session()

    def _headers(self) -> dict:
        return {
            "Authorization": f"Bearer {self.api_key}",
            "content-type": "application/json",
            "accept": "application/json",
        }

    def chat(self, body: dict, timeout: int = 90) -> requests.Response:
        url = f"{self.base_url}/v1/chat/completions"
        return self.session.post(
            url, headers=self._headers(), json=body, timeout=timeout
        )

    def chat_stream(self, body: dict, timeout: int = 120) -> list[dict]:
        body = {**body, "stream": True}
        url = f"{self.base_url}/v1/chat/completions"
        response = self.session.post(
            url, headers=self._headers(), json=body, timeout=timeout, stream=True
        )
        response.raise_for_status()
        events = []
        for line in response.iter_lines(decode_unicode=True):
            if not line:
                continue
            if line.startswith("data: "):
                data = line[6:]
                if data.strip() == "[DONE]":
                    break
                try:
                    events.append(json.loads(data))
                except json.JSONDecodeError:
                    pass
        return events


# ---------------------------------------------------------------------------
# Response Extraction Helpers
# ---------------------------------------------------------------------------

def extract_text_from_claude(data: dict) -> str:
    content = data.get("content", [])
    return "".join(
        b.get("text", "") for b in content
        if isinstance(b, dict) and b.get("type") == "text"
    )


def extract_text_from_openai(data: dict) -> str:
    for choice in data.get("choices", []):
        msg = choice.get("message", {})
        content = msg.get("content", "")
        if isinstance(content, str):
            return content
    return ""


def extract_tool_calls_claude(data: dict) -> list[dict]:
    return [
        {"name": b.get("name", ""), "input": b.get("input", {}), "id": b.get("id", "")}
        for b in data.get("content", [])
        if isinstance(b, dict) and b.get("type") == "tool_use"
    ]


def extract_tool_calls_openai(data: dict) -> list[dict]:
    tool_calls = []
    for choice in data.get("choices", []):
        for tc in choice.get("message", {}).get("tool_calls", []):
            fn = tc.get("function", {})
            args_str = fn.get("arguments", "{}")
            try:
                args = json.loads(args_str)
            except (json.JSONDecodeError, TypeError):
                args = args_str
            tool_calls.append({
                "id": tc.get("id", ""),
                "name": fn.get("name", ""),
                "arguments": args,
            })
    return tool_calls


def extract_usage_claude(data: dict) -> Optional[dict]:
    usage = data.get("usage", {})
    if not usage:
        return None
    return {
        "prompt_tokens": usage.get("input_tokens", 0),
        "completion_tokens": usage.get("output_tokens", 0),
        "total_tokens": usage.get("input_tokens", 0) + usage.get("output_tokens", 0),
    }


def extract_usage_openai(data: dict) -> Optional[dict]:
    usage = data.get("usage", {})
    if not usage:
        return None
    return {
        "prompt_tokens": usage.get("prompt_tokens", 0),
        "completion_tokens": usage.get("completion_tokens", 0),
        "total_tokens": usage.get("total_tokens", 0),
        "cache_hit_tokens": usage.get("prompt_tokens_details", {}).get("cached_tokens", 0),
        "reasoning_tokens": usage.get("completion_tokens_details", {}).get("reasoning_tokens", 0),
    }


def extract_stop_reason_claude(data: dict) -> Optional[str]:
    return data.get("stop_reason")


def extract_finish_reason_openai(data: dict) -> Optional[str]:
    choices = data.get("choices", [])
    return choices[0].get("finish_reason") if choices else None


def has_tool_result_claude(data: dict) -> bool:
    """Check if response contains tool_use blocks (not text)."""
    return any(
        isinstance(b, dict) and b.get("type") == "tool_use"
        for b in data.get("content", [])
    )


def has_tool_result_openai(data: dict) -> bool:
    """Check if response contains tool_calls (not text)."""
    for choice in data.get("choices", []):
        if choice.get("message", {}).get("tool_calls"):
            return True
    return False


# ---------------------------------------------------------------------------
# Billing / Token Verification
# ---------------------------------------------------------------------------

def verify_token_accuracy(
    uniapi_usage: Optional[dict],
    deepseek_usage: Optional[dict],
    prompt_tolerance: float = PROMPT_TOKEN_TOLERANCE,
) -> dict:
    """Verify billing/usage accuracy.

    Primary checks:
    - Prompt tokens must match closely (same input to same model).
    - Completion tokens must be > 0 (response was generated).
    - Total tokens must be consistent with prompt + completion.

    Completion tokens are NOT compared across calls because each API call
    produces different text (even with identical prompts). Instead we verify
    that UniAPI correctly reports non-zero completion tokens.
    """
    if not uniapi_usage or not deepseek_usage:
        return {"status": "skip", "reason": "Missing usage data"}

    result = {"status": "pass", "diffs": {}}
    issues = []

    # --- Prompt tokens: tight tolerance (same input) ---
    uni_p = uniapi_usage.get("prompt_tokens", 0)
    ds_p = deepseek_usage.get("prompt_tokens", 0)
    if ds_p > 0:
        diff_pct = abs(uni_p - ds_p) / ds_p
        result["diffs"]["Prompt"] = {
            "uniapi": uni_p, "deepseek": ds_p,
            "diff_pct": round(diff_pct * 100, 1),
            "tolerance": f"{prompt_tolerance*100:.0f}%",
        }
        if diff_pct > prompt_tolerance:
            issues.append(f"Prompt token mismatch: {uni_p} vs {ds_p}")
    elif uni_p > 0:
        result["diffs"]["Prompt"] = {"uniapi": uni_p, "deepseek": 0, "note": "DeepSeek returned 0"}

    # --- Completion tokens: verify non-zero (don't compare across calls) ---
    uni_c = uniapi_usage.get("completion_tokens", 0)
    ds_c = deepseek_usage.get("completion_tokens", 0)
    result["diffs"]["Completion"] = {
        "uniapi": uni_c, "deepseek": ds_c,
        "note": "Not compared — each call produces different text",
    }
    if uni_c == 0:
        issues.append("UniAPI reported 0 completion tokens")

    # --- Total tokens: verify consistency ---
    uni_t = uniapi_usage.get("total_tokens", 0)
    ds_t = deepseek_usage.get("total_tokens", 0)
    result["diffs"]["Total"] = {
        "uniapi": uni_t, "deepseek": ds_t,
    }

    # Verify UniAPI total = prompt + completion
    if uni_p + uni_c != uni_t:
        issues.append(f"UniAPI total mismatch: {uni_p}+{uni_c} != {uni_t}")

    # Cache hit tokens (if any)
    ds_cache = deepseek_usage.get("cache_hit_tokens", 0) or 0
    result["diffs"]["CacheHit"] = {"deepseek": ds_cache}

    if issues:
        result["status"] = "fail"
        result["issues"] = issues

    return result


# ---------------------------------------------------------------------------
# Test Runner
# ---------------------------------------------------------------------------

class TestRunner:
    def __init__(self, uniapi: UniAPIClient, deepseek: DeepSeekClient,
                 verbose: bool = False):
        self.uniapi = uniapi
        self.deepseek = deepseek
        self.verbose = verbose
        self.results: list[TestResult] = []

    def log(self, msg: str):
        if self.verbose:
            print(f"  {msg}")

    def run_test(
        self,
        name: str,
        category: str,
        model: str,
        claude_body: dict,
        openai_body: dict,
        has_tools: bool = False,
        check_stream: bool = False,
        check_billing: bool = True,
    ) -> TestResult:
        """Run a single test case."""
        start = time.monotonic()
        result = TestResult(name=name, category=category, model=model,
                            passed=False, duration_ms=0)

        try:
            # --- Non-streaming: UniAPI ---
            self.log(f"[{name}] UniAPI (Claude Messages)...")
            uni_resp = self.uniapi.chat(claude_body)
            result.uniapi_status = uni_resp.status_code

            if uni_resp.status_code >= 500:
                result.error = f"UniAPI server error: {uni_resp.status_code}"
                self.results.append(result)
                return result

            # --- Non-streaming: DeepSeek direct ---
            self.log(f"[{name}] DeepSeek direct (OpenAI)...")
            ds_resp = self.deepseek.chat(openai_body)
            result.deepseek_status = ds_resp.status_code

            if ds_resp.status_code >= 500:
                result.error = f"DeepSeek server error: {ds_resp.status_code}"
                self.results.append(result)
                return result

            uni_ok = 200 <= uni_resp.status_code < 300
            ds_ok = 200 <= ds_resp.status_code < 300

            if uni_ok and ds_ok:
                uni_data = uni_resp.json()
                ds_data = ds_resp.json()

                checks = {}

                # --- Stop reason ---
                uni_stop = extract_stop_reason_claude(uni_data)
                ds_stop = extract_finish_reason_openai(ds_data)
                check_stop_ok = stop_reasons_match(uni_stop, ds_stop)
                checks["stop_reason_match"] = check_stop_ok
                if not check_stop_ok:
                    checks["stop_reason_detail"] = f"Claude={uni_stop} OpenAI={ds_stop}"

                # --- Tool calls ---
                uni_tools = extract_tool_calls_claude(uni_data)
                ds_tools = extract_tool_calls_openai(ds_data)

                checks["tool_count_match"] = len(uni_tools) == len(ds_tools)

                if uni_tools and ds_tools:
                    uni_names = sorted(t.get("name", "") for t in uni_tools)
                    ds_names = sorted(t.get("name", "") for t in ds_tools)
                    checks["tool_names_match"] = uni_names == ds_names
                    if not checks["tool_names_match"]:
                        checks["tool_names_detail"] = f"Claude={uni_names} OpenAI={ds_names}"

                # --- Content presence ---
                if has_tools:
                    # For tool tests, either text OR tool_use is a valid response
                    uni_has_content = (
                        len(extract_text_from_claude(uni_data)) > 0 or
                        len(uni_tools) > 0
                    )
                    ds_has_content = (
                        len(extract_text_from_openai(ds_data)) > 0 or
                        len(ds_tools) > 0
                    )
                else:
                    uni_has_content = len(extract_text_from_claude(uni_data)) > 0
                    ds_has_content = len(extract_text_from_openai(ds_data)) > 0

                checks["response_has_content_uni"] = uni_has_content
                checks["response_has_content_ds"] = ds_has_content

                result.details = checks

                # --- Billing / Usage ---
                if check_billing:
                    uni_usage = extract_usage_claude(uni_data)
                    ds_usage = extract_usage_openai(ds_data)
                    result.uniapi_tokens = uni_usage
                    result.deepseek_tokens = ds_usage
                    billing = verify_token_accuracy(uni_usage, ds_usage)
                    result.billing_check = billing["status"]
                    result.details["billing"] = billing

                # --- Overall pass ---
                essential_checks = ["stop_reason_match", "response_has_content_uni",
                                    "response_has_content_ds"]
                if uni_tools or ds_tools:
                    essential_checks.append("tool_count_match")
                all_essential = all(checks.get(c, False) for c in essential_checks)
                billing_ok = result.billing_check in ("pass", "skip", None)
                result.passed = all_essential and billing_ok

            elif not uni_ok and not ds_ok:
                result.details["both_errored"] = True
                result.details["status_codes"] = (
                    f"UniAPI={uni_resp.status_code}, DeepSeek={ds_resp.status_code}"
                )
                result.passed = True
            else:
                result.details["status_mismatch"] = (
                    f"UniAPI={uni_resp.status_code}, DeepSeek={ds_resp.status_code}"
                )
                # If UniAPI succeeded but DeepSeek failed, could be a conversion issue
                try:
                    uni_err = uni_resp.json() if uni_ok else None
                    ds_err = ds_resp.json() if not ds_ok else None
                    result.details["uni_response"] = str(uni_err)[:500]
                    result.details["ds_response"] = str(ds_err)[:500]
                except Exception:
                    pass
                result.passed = False

            # --- Streaming ---
            if check_stream and result.passed:
                self.log(f"[{name}] Streaming...")
                self._check_streaming(name, claude_body, openai_body, has_tools, result)

        except requests.exceptions.Timeout:
            result.error = "Request timeout"
        except requests.exceptions.ConnectionError as e:
            result.error = f"Connection error: {e}"
        except Exception as e:
            result.error = f"Unexpected error: {e}\n{traceback.format_exc()}"

        result.duration_ms = round((time.monotonic() - start) * 1000, 1)
        self.results.append(result)
        return result

    def _check_streaming(
        self, name: str,
        claude_body: dict, openai_body: dict,
        has_tools: bool, result: TestResult,
    ):
        """Verify streaming works for both UniAPI and DeepSeek."""
        try:
            uni_events = self.uniapi.chat_stream(dict(claude_body))
            ds_events = self.deepseek.chat_stream(dict(openai_body))

            # Collect UniAPI stream text
            uni_text = ""
            for evt in uni_events:
                t = evt.get("type", "")
                if t == "content_block_delta":
                    delta = evt.get("delta", {})
                    if isinstance(delta, dict) and delta.get("type") == "text_delta":
                        uni_text += delta.get("text", "")
                elif t == "content_block_delta":
                    delta = evt.get("delta", {})
                    if isinstance(delta, dict) and delta.get("type") == "input_json_delta":
                        uni_text += delta.get("partial_json", "")

            # Collect DeepSeek stream text
            ds_text = ""
            for chunk in ds_events:
                for c in chunk.get("choices", []):
                    delta = c.get("delta", {})
                    content = delta.get("content", "")
                    if isinstance(content, str):
                        ds_text += content
                    # Tool call deltas
                    for tc in delta.get("tool_calls", []):
                        fn = tc.get("function", {})
                        ds_text += fn.get("arguments", "")

            result.details["stream_uni_events"] = len(uni_events)
            result.details["stream_ds_events"] = len(ds_events)
            result.details["stream_uni_text_len"] = len(uni_text)
            result.details["stream_ds_text_len"] = len(ds_text)

            if len(uni_events) == 0:
                result.details["stream_uni_empty"] = True
                result.passed = False

        except Exception as e:
            result.details["stream_error"] = str(e)


# ---------------------------------------------------------------------------
# Test Case Definitions
# ---------------------------------------------------------------------------

def get_test_cases(model: str) -> list[dict]:
    """Build all test cases for a given model name."""

    WEATHER_TOOL_CLAUDE = [{
        "name": "get_weather",
        "description": "Get the current weather for a city",
        "input_schema": {
            "type": "object",
            "properties": {"city": {"type": "string", "description": "City name"}},
            "required": ["city"],
        },
    }]

    WEATHER_TOOL_OPENAI = [{
        "type": "function",
        "function": {
            "name": "get_weather",
            "description": "Get the current weather for a city",
            "parameters": {
                "type": "object",
                "properties": {"city": {"type": "string", "description": "City name"}},
                "required": ["city"],
            },
        },
    }]

    MULTI_TOOLS_CLAUDE = [
        {
            "name": "get_weather",
            "description": "Get weather for a city",
            "input_schema": {
                "type": "object",
                "properties": {"city": {"type": "string"}},
                "required": ["city"],
            },
        },
        {
            "name": "get_time",
            "description": "Get current time for a timezone",
            "input_schema": {
                "type": "object",
                "properties": {"timezone": {"type": "string"}},
                "required": ["timezone"],
            },
        },
    ]

    MULTI_TOOLS_OPENAI = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get weather for a city",
                "parameters": {
                    "type": "object",
                    "properties": {"city": {"type": "string"}},
                    "required": ["city"],
                },
            },
        },
        {
            "type": "function",
            "function": {
                "name": "get_time",
                "description": "Get current time for a timezone",
                "parameters": {
                    "type": "object",
                    "properties": {"timezone": {"type": "string"}},
                    "required": ["timezone"],
                },
            },
        },
    ]

    return [
        # ============================================================
        # BASIC CONVERSATION
        # ============================================================
        {
            "name": "basic_simple_query",
            "category": "basic",
            "check_stream": True,
            "claude": make_claude_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Say hello in exactly 3 words."}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Say hello in exactly 3 words."}],
            ),
        },
        {
            "name": "basic_system_prompt",
            "category": "basic",
            "claude": make_claude_request(
                model, max_tokens=256,
                system="You are a helpful assistant. Always answer in JSON format.",
                messages=[{"role": "user", "content": 'What is 2+2? Reply with {"answer": N}.'}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                messages=[
                    {"role": "system", "content": "You are a helpful assistant. Always answer in JSON format."},
                    {"role": "user", "content": 'What is 2+2? Reply with {"answer": N}.'},
                ],
            ),
        },
        {
            "name": "basic_multi_turn",
            "category": "basic",
            "check_stream": True,
            "claude": make_claude_request(
                model, max_tokens=256,
                messages=[
                    {"role": "user", "content": "My name is Alice."},
                    {"role": "assistant", "content": "Hello Alice! How can I help you today?"},
                    {"role": "user", "content": "What is my name?"},
                ],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                messages=[
                    {"role": "user", "content": "My name is Alice."},
                    {"role": "assistant", "content": "Hello Alice! How can I help you today?"},
                    {"role": "user", "content": "What is my name?"},
                ],
            ),
        },

        # ============================================================
        # THINKING MODE
        # ============================================================
        {
            "name": "thinking_enabled",
            "category": "thinking",
            "claude": make_claude_request(
                model, max_tokens=512,
                thinking={"type": "enabled", "budget_tokens": 2048},
                messages=[{"role": "user", "content": "Solve step by step: If a train travels 120 miles in 2 hours, what is its average speed?"}],
            ),
            "openai": make_openai_request(
                model, max_tokens=512,
                messages=[{"role": "user", "content": "Solve step by step: If a train travels 120 miles in 2 hours, what is its average speed?"}],
            ),
        },
        {
            "name": "thinking_disabled",
            "category": "thinking",
            "claude": make_claude_request(
                model, max_tokens=256,
                thinking={"type": "disabled"},
                messages=[{"role": "user", "content": "Say hello and nothing else."}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Say hello and nothing else."}],
            ),
        },
        {
            "name": "thinking_default",
            "category": "thinking",
            "claude": make_claude_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Count from 1 to 5."}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Count from 1 to 5."}],
            ),
        },

        # ============================================================
        # TOOL CALLING
        # ============================================================
        {
            "name": "tool_single_call",
            "category": "tool_calling",
            "has_tools": True,
            "claude": make_claude_request(
                model, max_tokens=256,
                tools=WEATHER_TOOL_CLAUDE,
                messages=[{"role": "user", "content": "What is the weather in Tokyo?"}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                tools=WEATHER_TOOL_OPENAI,
                messages=[{"role": "user", "content": "What is the weather in Tokyo?"}],
            ),
        },
        {
            "name": "tool_multi_call",
            "category": "tool_calling",
            "has_tools": True,
            "claude": make_claude_request(
                model, max_tokens=256,
                tools=MULTI_TOOLS_CLAUDE,
                messages=[{"role": "user", "content": "What time is it in London and what is the weather there?"}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                tools=MULTI_TOOLS_OPENAI,
                messages=[{"role": "user", "content": "What time is it in London and what is the weather there?"}],
            ),
        },
        {
            "name": "tool_result_roundtrip",
            "category": "tool_calling",
            "has_tools": True,
            # NOTE: DeepSeek V4 enables thinking mode by default and requires
            # reasoning_content on every assistant message. UniAPI automatically
            # injects missing reasoning_content. The direct OpenAI request must
            # include it manually — this test validates UniAPI's injection.
            "claude": make_claude_request(
                model, max_tokens=256,
                tools=WEATHER_TOOL_CLAUDE,
                messages=[
                    {"role": "user", "content": "What is the weather in Tokyo?"},
                    {"role": "assistant", "content": [
                        {"type": "tool_use", "id": "toolu_001", "name": "get_weather",
                         "input": {"city": "Tokyo"}},
                    ]},
                    {"role": "user", "content": [
                        {"type": "tool_result", "tool_use_id": "toolu_001",
                         "content": "Tokyo weather: 22C, partly cloudy"},
                    ]},
                ],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                tools=WEATHER_TOOL_OPENAI,
                messages=[
                    {"role": "user", "content": "What is the weather in Tokyo?"},
                    {"role": "assistant", "content": None,
                     "reasoning_content": "",  # Required by DeepSeek V4 thinking mode
                     "tool_calls": [
                         {"id": "toolu_001", "type": "function",
                          "function": {"name": "get_weather", "arguments": '{"city": "Tokyo"}'}},
                     ]},
                    {"role": "tool", "tool_call_id": "toolu_001",
                     "name": "get_weather",
                     "content": "Tokyo weather: 22C, partly cloudy"},
                ],
            ),
        },
        {
            "name": "tool_with_thinking",
            "category": "tool_calling",
            "has_tools": True,
            "claude": make_claude_request(
                model, max_tokens=512,
                thinking={"type": "enabled", "budget_tokens": 2048},
                tools=WEATHER_TOOL_CLAUDE,
                messages=[{"role": "user", "content": "What is the weather in Paris? If rainy, suggest indoor activities."}],
            ),
            "openai": make_openai_request(
                model, max_tokens=512,
                tools=WEATHER_TOOL_OPENAI,
                messages=[{"role": "user", "content": "What is the weather in Paris? If rainy, suggest indoor activities."}],
            ),
        },

        # ============================================================
        # EDGE CASES
        # ============================================================
        {
            "name": "edge_unicode_emoji",
            "category": "edge_cases",
            "claude": make_claude_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Translate to Chinese: Hello! How are you? 😊"}],
            ),
            "openai": make_openai_request(
                model, max_tokens=256,
                messages=[{"role": "user", "content": "Translate to Chinese: Hello! How are you? 😊"}],
            ),
        },
        {
            "name": "edge_code_generation",
            "category": "edge_cases",
            # DeepSeek V4 enables thinking mode by default — reasoning tokens
            # consume the max_tokens budget. Use a larger limit so the full
            # code output fits after reasoning.
            "claude": make_claude_request(
                model, max_tokens=1024,
                messages=[{"role": "user", "content": "Write a Python function that calculates fibonacci numbers recursively. Include a docstring."}],
            ),
            "openai": make_openai_request(
                model, max_tokens=1024,
                messages=[{"role": "user", "content": "Write a Python function that calculates fibonacci numbers recursively. Include a docstring."}],
            ),
        },

        # ============================================================
        # ERROR CASES
        # ============================================================
        {
            "name": "error_invalid_model",
            "category": "error_cases",
            "check_billing": False,
            "claude": make_claude_request(
                "nonexistent-model-12345", max_tokens=10,
                messages=[{"role": "user", "content": "Hi"}],
            ),
            "openai": make_openai_request(
                "nonexistent-model-12345", max_tokens=10,
                messages=[{"role": "user", "content": "Hi"}],
            ),
        },
    ]


# ---------------------------------------------------------------------------
# Report Generator
# ---------------------------------------------------------------------------

def print_report(results: list[TestResult], model: str):
    passed = sum(1 for r in results if r.passed)
    failed = sum(1 for r in results if not r.passed)
    total = len(results)

    print()
    print("=" * 78)
    print(f"  DeepSeek Compatibility Test Report")
    print(f"  Model: {model}")
    print(f"  Time:  {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M:%S UTC')}")
    print(f"  Total: {total} | Passed: {passed} | Failed: {failed} | "
          f"Pass rate: {passed/total*100:.0f}%" if total else "")
    print("=" * 78)

    categories = {}
    for r in results:
        categories.setdefault(r.category, []).append(r)

    for cat, cat_results in categories.items():
        cat_passed = sum(1 for r in cat_results if r.passed)
        cat_total = len(cat_results)
        print(f"\n── {cat.upper()} ({cat_passed}/{cat_total}) ──")
        for r in cat_results:
            status = "PASS" if r.passed else "FAIL"
            print(f"  [{status}] {r.name}  ({r.duration_ms}ms)")
            if r.uniapi_status:
                print(f"      HTTP: UniAPI={r.uniapi_status}  DeepSeek={r.deepseek_status}")
            if r.uniapi_tokens and r.deepseek_tokens:
                u = r.uniapi_tokens
                d = r.deepseek_tokens
                print(f"      Tokens UniAPI:    p={u['prompt_tokens']:>5}  c={u['completion_tokens']:>5}  t={u['total_tokens']:>5}")
                print(f"      Tokens DeepSeek:  p={d['prompt_tokens']:>5}  c={d['completion_tokens']:>5}  t={d['total_tokens']:>5}")
            if r.billing_check:
                icon = "PASS" if r.billing_check == "pass" else ("SKIP" if r.billing_check == "skip" else "FAIL")
                print(f"      Billing: {icon}")
                billing = r.details.get("billing", {})
                for label, diff in billing.get("diffs", {}).items():
                    if isinstance(diff, dict):
                        t = diff.get("tolerance", "")
                        note = diff.get("note", "")
                        uni_v = diff.get("uniapi", "N/A")
                        ds_v = diff.get("deepseek", "N/A")
                        if "diff_pct" in diff:
                            print(f"        {label}: UniAPI={uni_v} DeepSeek={ds_v} "
                                  f"diff={diff['diff_pct']}% (limit={t})")
                        elif note:
                            print(f"        {label}: UniAPI={uni_v} DeepSeek={ds_v} ({note})")
                        else:
                            print(f"        {label}: UniAPI={uni_v} DeepSeek={ds_v}")
            # Show failure details
            for k, v in r.details.items():
                if k in ("billing",):
                    continue
                if isinstance(v, bool) and not v:
                    print(f"      {k}: FAIL")
                elif isinstance(v, str) and k.endswith("_detail"):
                    print(f"      {k}: {v}")
                elif isinstance(v, str) and k.endswith("_mismatch"):
                    print(f"      {k}: {v}")
            if r.error:
                print(f"      Error: {r.error}")

    print()
    print("=" * 78)
    if failed == 0:
        print("  RESULT: ALL TESTS PASSED")
    else:
        print(f"  RESULT: {failed}/{total} TESTS FAILED")
        print()
        print("  FAILURE SUMMARY:")
        for r in results:
            if not r.passed:
                reasons = []
                if r.error:
                    reasons.append(f"error={r.error}")
                for k, v in r.details.items():
                    if k in ("billing",):
                        continue
                    if isinstance(v, bool) and not v:
                        reasons.append(k)
                    elif isinstance(v, str) and (k.endswith("_detail") or k.endswith("_mismatch")):
                        reasons.append(f"{k}={v}")
                print(f"    - {r.name}: {', '.join(reasons) if reasons else 'see details above'}")
    print("=" * 78)
    print()

    return failed == 0


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(
        description="DeepSeek Compatibility Test Suite (Claude Messages via UniAPI)"
    )
    parser.add_argument("--model", choices=MODELS, default=None,
                        help="Test a specific model (default: test all)")
    parser.add_argument("--verbose", "-v", action="store_true",
                        help="Verbose output")
    parser.add_argument("--dry-run", action="store_true",
                        help="Print test cases without executing")
    args = parser.parse_args()

    models = [args.model] if args.model else MODELS

    if not args.dry_run:
        if UNIAPI_API_KEY == "sk-test-key":
            print("WARNING: UNIAPI_API_KEY is not set. Export it to run tests.")
        if DEEPSEEK_API_KEY == "sk-your-deepseek-key":
            print("WARNING: DEEPSEEK_API_KEY is not set. Export it to run tests.")

    uniapi = UniAPIClient(UNIAPI_BASE_URL, UNIAPI_API_KEY)
    deepseek = DeepSeekClient(DEEPSEEK_API_KEY)

    all_passed = True

    for model in models:
        print(f"\n{'─' * 60}")
        print(f"Testing model: {model}")
        print(f"UniAPI: {UNIAPI_BASE_URL}")
        print(f"{'─' * 60}")

        runner = TestRunner(uniapi, deepseek, verbose=args.verbose)
        cases = get_test_cases(model)

        if args.dry_run:
            print(f"\nDry run — {len(cases)} test cases:\n")
            for i, case in enumerate(cases):
                print(f"  {i+1}. [{case['category']}] {case['name']}")
                print(f"     Claude keys: {list(case['claude'].keys())}")
                print(f"     OpenAI keys: {list(case['openai'].keys())}")
                print(f"     Tools: {case.get('has_tools', False)}")
                print(f"     Stream: {case.get('check_stream', False)}")
                print(f"     Billing: {case.get('check_billing', True)}")
            continue

        for case in cases:
            check_stream = case.pop("check_stream", False)
            check_billing = case.pop("check_billing", True)
            has_tools = case.pop("has_tools", False)
            result = runner.run_test(
                name=case["name"],
                category=case["category"],
                model=model,
                claude_body=case["claude"],
                openai_body=case["openai"],
                has_tools=has_tools,
                check_stream=check_stream,
                check_billing=check_billing,
            )
            status = "PASS" if result.passed else "FAIL"
            print(f"  [{status}] {result.name} ({result.duration_ms}ms)")

        if not args.dry_run:
            model_passed = print_report(runner.results, model)
            all_passed = all_passed and model_passed

    sys.exit(0 if all_passed else 1)


if __name__ == "__main__":
    main()
