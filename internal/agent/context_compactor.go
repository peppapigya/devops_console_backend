package agent

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	maxExcerptChars   = 1200
	maxEvidenceLines  = 8
	maxSummaryPreview = 280
)

var evidenceKeywords = []string{
	"error", "failed", "failure", "emerg", "warn", "critical", "syntax",
	"unexpected", "not terminated", "invalid", "denied", "panic", "traceback",
	"line ", "column ", "nginx:", "config", "refused", "timeout",
}

var lineNumberPattern = regexp.MustCompile(`:[0-9]+`)

func CompactToolResult(toolName, command string, result *ToolCallResult) map[string]interface{} {
	output := strings.TrimSpace(result.Output)
	stderr := strings.TrimSpace(result.Stderr)
	combined := strings.TrimSpace(strings.Join([]string{output, stderr}, "\n"))

	evidence := extractEvidenceLines(combined, maxEvidenceLines)
	summary := buildResultSummary(result.Success, result.ExitCode, evidence, combined)
	excerpt, truncated := buildExcerpt(combined, maxExcerptChars)

	return map[string]interface{}{
		"tool_name":        toolName,
		"command":          command,
		"success":          result.Success,
		"exit_code":        result.ExitCode,
		"duration_ms":      result.DurationMs,
		"summary":          summary,
		"key_evidence":     evidence,
		"output_excerpt":   excerpt,
		"truncated":        truncated,
		"raw_output_chars": len(output),
		"raw_stderr_chars": len(stderr),
	}
}

func extractEvidenceLines(text string, limit int) []string {
	if text == "" || limit <= 0 {
		return nil
	}
	lines := strings.Split(text, "\n")
	seen := map[string]struct{}{}
	picked := make([]string, 0, limit)

	appendIfUseful := func(line string) {
		line = strings.TrimSpace(line)
		if line == "" {
			return
		}
		if len(line) > 260 {
			line = line[:260] + "..."
		}
		if _, ok := seen[line]; ok {
			return
		}
		picked = append(picked, line)
		seen[line] = struct{}{}
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		lower := strings.ToLower(line)
		matched := false
		for _, kw := range evidenceKeywords {
			if strings.Contains(lower, kw) {
				matched = true
				break
			}
		}
		if !matched && !lineNumberPattern.MatchString(lower) {
			continue
		}
		appendIfUseful(line)
		if len(picked) >= limit {
			return picked
		}
	}

	// fallback: keep the first few non-empty lines
	for _, raw := range lines {
		appendIfUseful(raw)
		if len(picked) >= limit {
			break
		}
	}
	return picked
}

func buildExcerpt(text string, limit int) (string, bool) {
	if limit <= 0 || len(text) <= limit {
		return text, false
	}
	head := limit / 2
	tail := limit - head
	return text[:head] + "\n...\n" + text[len(text)-tail:], true
}

func buildResultSummary(success bool, exitCode int, evidence []string, combined string) string {
	status := "执行失败"
	if success && exitCode == 0 {
		status = "执行成功"
	}
	if len(evidence) > 0 {
		return fmt.Sprintf("%s，关键证据：%s", status, strings.Join(evidence[:minInt(2, len(evidence))], " | "))
	}
	if combined != "" {
		preview := combined
		if len(preview) > maxSummaryPreview {
			preview = preview[:maxSummaryPreview] + "..."
		}
		return fmt.Sprintf("%s，输出摘要：%s", status, preview)
	}
	return status
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
