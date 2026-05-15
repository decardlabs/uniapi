package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Laisky/errors/v2"
	glog "github.com/Laisky/go-utils/v6/log"
	"github.com/Laisky/zap"
)

var responseIDRegexp = regexp.MustCompile(`"id"\s*:\s*"([^"]+)"`)

// liveOptions defines command-line options for the real-channel live probe command.
type liveOptions struct {
	model       string
	rounds      int
	concurrency int
	timeout     time.Duration
}

// liveStepResult records one validation step outcome inside a live probe scenario.
type liveStepResult struct {
	Name       string
	Successful bool
	Skipped    bool
	StatusCode int
	Reason     string
}

// liveScenarioResult captures all step results for one scenario round.
type liveScenarioResult struct {
	Round int
	Steps []liveStepResult
}

// liveSummary aggregates scenario results across all rounds.
type liveSummary struct {
	totalRounds  int
	passedRounds int
	stepStats    map[string]liveStepStat
	failures     []liveScenarioResult
}

// liveStepStat stores aggregate counters for a single validation step.
type liveStepStat struct {
	Total   int
	Passed  int
	Skipped int
}

// live executes real-channel usage probes against configured models and endpoints.
//
// Parameters:
//   - ctx: command context for cancellation.
//   - logger: structured logger for diagnostics.
//   - args: command-line arguments passed after `live`.
//
// Returns:
//   - error: non-nil when configuration, execution, or summary validation fails.
func live(ctx context.Context, logger glog.Logger, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return errors.Wrap(err, "load config")
	}

	opts, err := parseLiveArgs(args, cfg)
	if err != nil {
		return errors.Wrap(err, "parse live arguments")
	}

	logger.Info("starting live channel probe",
		zap.String("base_url", cfg.APIBase),
		zap.String("model", opts.model),
		zap.Int("rounds", opts.rounds),
		zap.Int("concurrency", opts.concurrency),
		zap.Duration("timeout", opts.timeout),
	)

	httpClient := &http.Client{Timeout: opts.timeout}
	resultsCh := make(chan liveScenarioResult, opts.rounds)
	jobs := make(chan int, opts.rounds)

	for round := 1; round <= opts.rounds; round++ {
		jobs <- round
	}
	close(jobs)

	var workers sync.WaitGroup
	for worker := 0; worker < opts.concurrency; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for round := range jobs {
				roundResult := runLiveScenario(ctx, httpClient, cfg, opts, round)
				select {
				case resultsCh <- roundResult:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		workers.Wait()
		close(resultsCh)
	}()

	summary := liveSummary{stepStats: make(map[string]liveStepStat, 4)}
	for scenario := range resultsCh {
		summary.totalRounds++
		allPassed := true
		for _, step := range scenario.Steps {
			stat := summary.stepStats[step.Name]
			if step.Skipped {
				stat.Skipped++
				summary.stepStats[step.Name] = stat
				continue
			}
			stat.Total++
			if step.Successful {
				stat.Passed++
			} else {
				allPassed = false
			}
			summary.stepStats[step.Name] = stat
		}

		if allPassed {
			summary.passedRounds++
		} else {
			summary.failures = append(summary.failures, scenario)
		}
	}

	renderLiveSummary(logger, summary)
	if summary.passedRounds != summary.totalRounds {
		return errors.Errorf("live probe failed: %d/%d rounds passed", summary.passedRounds, summary.totalRounds)
	}

	return nil
}

// parseLiveArgs parses command-line flags for live probes.
//
// Parameters:
//   - args: raw CLI arguments after the `live` token.
//   - cfg: shared test harness configuration for defaults.
//
// Returns:
//   - liveOptions: validated command options.
//   - error: non-nil when arguments are invalid.
func parseLiveArgs(args []string, cfg config) (liveOptions, error) {
	fs := flag.NewFlagSet("live", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	defaultModel := ""
	if len(cfg.Models) > 0 {
		defaultModel = strings.TrimSpace(cfg.Models[0])
	}

	var opts liveOptions
	fs.StringVar(&opts.model, "model", defaultModel, "model name used for live probes")
	fs.IntVar(&opts.rounds, "rounds", 3, "number of rounds to execute")
	fs.IntVar(&opts.concurrency, "concurrency", 1, "number of concurrent workers")
	fs.DurationVar(&opts.timeout, "timeout", 90*time.Second, "per-request timeout")

	if err := fs.Parse(args); err != nil {
		return liveOptions{}, errors.Wrap(err, "parse live flags")
	}

	opts.model = strings.TrimSpace(opts.model)
	if opts.model == "" {
		return liveOptions{}, errors.New("model is required; set --model or ONEAPI_TEST_MODELS")
	}
	if opts.rounds <= 0 {
		return liveOptions{}, errors.New("rounds must be positive")
	}
	if opts.concurrency <= 0 {
		return liveOptions{}, errors.New("concurrency must be positive")
	}
	if opts.timeout <= 0 {
		return liveOptions{}, errors.New("timeout must be positive")
	}

	return opts, nil
}

// runLiveScenario executes one round of end-to-end checks for practical usage behavior.
//
// Parameters:
//   - ctx: command context for cancellation.
//   - client: HTTP client used for requests.
//   - cfg: runtime endpoint and token configuration.
//   - opts: live probe options.
//   - round: 1-based round number.
//
// Returns:
//   - liveScenarioResult: step-by-step outcomes for this round.
func runLiveScenario(ctx context.Context, client *http.Client, cfg config, opts liveOptions, round int) liveScenarioResult {
	result := liveScenarioResult{Round: round}

	requestCtx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	baseInput := fmt.Sprintf("Round %d: summarize one practical tip about load testing in exactly one sentence.", round)

	createPayload := responseAPIPayload(opts.model, false, expectationDefault)
	createBody, ok := createPayload.(map[string]any)
	if !ok {
		result.Steps = append(result.Steps, liveStepResult{
			Name:       "response_create",
			Successful: false,
			Reason:     "invalid response payload type",
		})
		return result
	}
	createBody["input"] = baseInput

	createSpec := requestSpec{
		RequestFormat: "live_response_create",
		Label:         "Live Response Create",
		Type:          requestTypeResponseAPI,
		Path:          "/v1/responses",
		Body:          createBody,
		Stream:        false,
		Expectation:   expectationDefault,
	}
	createRes := performRequest(requestCtx, client, cfg.APIBase, cfg.Token, createSpec, opts.model)
	result.Steps = append(result.Steps, liveStepResult{
		Name:       "response_create",
		Successful: createRes.Success,
		Skipped:    false,
		StatusCode: createRes.StatusCode,
		Reason:     nonSuccessReason(createRes),
	})

	responseID := extractResponseID(createRes.ResponseBody)
	if !createRes.Success || responseID == "" {
		reason := ""
		if !createRes.Success {
			reason = nonSuccessReason(createRes)
		} else {
			reason = "response id missing from response body"
		}
		result.Steps = append(result.Steps, liveStepResult{
			Name:       "response_continue",
			Successful: false,
			Skipped:    false,
			Reason:     reason,
		})
	} else {
		continuePayload := responseAPIPayload(opts.model, false, expectationDefault)
		continueBody, bodyOK := continuePayload.(map[string]any)
		if !bodyOK {
			result.Steps = append(result.Steps, liveStepResult{
				Name:       "response_continue",
				Successful: false,
				Skipped:    false,
				Reason:     "invalid continuation payload type",
			})
		} else {
			continueBody["input"] = "Continue with one extra sentence that references the previous answer."
			continueBody["previous_response_id"] = responseID
			continueSpec := requestSpec{
				RequestFormat: "live_response_continue",
				Label:         "Live Response Continue",
				Type:          requestTypeResponseAPI,
				Path:          "/v1/responses",
				Body:          continueBody,
				Stream:        false,
				Expectation:   expectationDefault,
			}
			continueRes := performRequest(requestCtx, client, cfg.APIBase, cfg.Token, continueSpec, opts.model)
			result.Steps = append(result.Steps, liveStepResult{
				Name:       "response_continue",
				Successful: continueRes.Success,
				Skipped:    false,
				StatusCode: continueRes.StatusCode,
				Reason:     nonSuccessReason(continueRes),
			})
		}
	}

	chatSpec := requestSpec{
		RequestFormat: "live_chat_completion",
		Label:         "Live Chat Completion",
		Type:          requestTypeChatCompletion,
		Path:          "/v1/chat/completions",
		Body:          chatCompletionPayload(opts.model, false, expectationDefault),
		Stream:        false,
		Expectation:   expectationDefault,
	}
	chatRes := performRequest(requestCtx, client, cfg.APIBase, cfg.Token, chatSpec, opts.model)
	result.Steps = append(result.Steps, liveStepResult{
		Name:       "chat_completion",
		Successful: chatRes.Success,
		Skipped:    false,
		StatusCode: chatRes.StatusCode,
		Reason:     nonSuccessReason(chatRes),
	})

	if liveSupportsToolInvocation(opts.model) {
		toolSpec := requestSpec{
			RequestFormat: "live_chat_tool",
			Label:         "Live Chat Tool Invocation",
			Type:          requestTypeChatCompletion,
			Path:          "/v1/chat/completions",
			Body:          chatCompletionPayload(opts.model, false, expectationToolInvocation),
			AttemptBodies: toolAttemptPayloads(requestTypeChatCompletion, opts.model, false, expectationToolInvocation),
			Stream:        false,
			Expectation:   expectationToolInvocation,
		}
		if len(toolSpec.AttemptBodies) > 0 {
			toolSpec.Body = toolSpec.AttemptBodies[0]
		}
		toolRes := performRequest(requestCtx, client, cfg.APIBase, cfg.Token, toolSpec, opts.model)
		result.Steps = append(result.Steps, liveStepResult{
			Name:       "chat_tool_invocation",
			Successful: toolRes.Success,
			Skipped:    false,
			StatusCode: toolRes.StatusCode,
			Reason:     nonSuccessReason(toolRes),
		})
	} else {
		result.Steps = append(result.Steps, liveStepResult{
			Name:       "chat_tool_invocation",
			Successful: false,
			Skipped:    true,
			Reason:     "skipped for model that does not reliably support forced tool_choice",
		})
	}

	return result
}

// renderLiveSummary logs final pass/fail metrics and compact failure details.
//
// Parameters:
//   - logger: structured logger for output.
//   - summary: aggregated summary to render.
func renderLiveSummary(logger glog.Logger, summary liveSummary) {
	stepNames := []string{"response_create", "response_continue", "chat_completion", "chat_tool_invocation"}
	for _, stepName := range stepNames {
		stat := summary.stepStats[stepName]
		rate := 0.0
		if stat.Total > 0 {
			rate = float64(stat.Passed) / float64(stat.Total) * 100
		}
		logger.Info("live probe step metrics",
			zap.String("step", stepName),
			zap.Int("passed", stat.Passed),
			zap.Int("total", stat.Total),
			zap.Int("skipped", stat.Skipped),
			zap.Float64("success_rate", rate),
		)
	}

	overallRate := 0.0
	if summary.totalRounds > 0 {
		overallRate = float64(summary.passedRounds) / float64(summary.totalRounds) * 100
	}
	logger.Info("live probe summary",
		zap.Int("passed_rounds", summary.passedRounds),
		zap.Int("total_rounds", summary.totalRounds),
		zap.Float64("overall_success_rate", overallRate),
	)

	maxFailureLogs := 3
	for idx, failure := range summary.failures {
		if idx >= maxFailureLogs {
			break
		}
		for _, step := range failure.Steps {
			if step.Successful {
				continue
			}
			logger.Warn("live probe failure sample",
				zap.Int("round", failure.Round),
				zap.String("step", step.Name),
				zap.Int("status", step.StatusCode),
				zap.String("reason", step.Reason),
			)
		}
	}
}

// liveSupportsToolInvocation reports whether the live probe should exercise forced tool calling for the model.
// The function returns false for DeepSeek models because the current harness's forced tool_choice flow is not
// a reliable compatibility signal for those models.
func liveSupportsToolInvocation(model string) bool {
	lower := strings.ToLower(strings.TrimSpace(model))
	if lower == "" {
		return true
	}
	if strings.HasPrefix(lower, "deepseek") {
		return false
	}
	return true
}

// nonSuccessReason normalizes failed result reasons for concise summary logging.
//
// Parameters:
//   - res: request-level probe result.
//
// Returns:
//   - string: normalized reason text when a request did not succeed.
func nonSuccessReason(res testResult) string {
	if res.Success {
		return ""
	}
	if strings.TrimSpace(res.ErrorReason) != "" {
		return res.ErrorReason
	}
	if strings.TrimSpace(res.Warning) != "" {
		return res.Warning
	}
	if res.Skipped {
		return "skipped"
	}
	return "unknown failure"
}

// extractResponseID parses the top-level response id from JSON text.
//
// Parameters:
//   - body: response body text from /v1/responses.
//
// Returns:
//   - string: parsed response id, or empty when not found.
func extractResponseID(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}

	matches := responseIDRegexp.FindStringSubmatch(trimmed)
	if len(matches) != 2 {
		return ""
	}

	candidate := strings.TrimSpace(matches[1])
	if candidate == "" {
		return ""
	}
	return candidate
}
