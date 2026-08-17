package openai_compat

import "testing"

func TestResolveResponsesSupport(t *testing.T) {
	tests := []struct {
		name  string
		extra map[string]any
		want  AccountResponsesSupport
	}{
		{"nil extra", nil, ResponsesSupportUnknown},
		{"empty extra", map[string]any{}, ResponsesSupportUnknown},
		{"explicit true", map[string]any{ExtraKeyResponsesSupported: true}, ResponsesSupportYes},
		{"explicit false", map[string]any{ExtraKeyResponsesSupported: false}, ResponsesSupportNo},
		{"legacy explicit true", map[string]any{ExtraKeyLegacyUseResponsesAPI: true}, ResponsesSupportYes},
		{"legacy explicit false", map[string]any{ExtraKeyLegacyUseResponsesAPI: false}, ResponsesSupportNo},
		{"wrong explicit type", map[string]any{ExtraKeyResponsesSupported: "true"}, ResponsesSupportUnknown},
		{"verified probe is advisory", map[string]any{ExtraKeyResponsesProbeStatus: string(ResponsesProbeStatusVerified)}, ResponsesSupportUnknown},
		{"unsupported probe is advisory", map[string]any{ExtraKeyResponsesProbeStatus: string(ResponsesProbeStatusUnsupported)}, ResponsesSupportUnknown},
		{"force responses", map[string]any{ExtraKeyResponsesMode: string(ResponsesSupportModeForceResponses)}, ResponsesSupportYes},
		{"force chat completions", map[string]any{ExtraKeyResponsesMode: string(ResponsesSupportModeForceChatCompletions)}, ResponsesSupportNo},
		{"force responses overrides explicit false", map[string]any{ExtraKeyResponsesMode: string(ResponsesSupportModeForceResponses), ExtraKeyResponsesSupported: false}, ResponsesSupportYes},
		{"force chat overrides explicit true", map[string]any{ExtraKeyResponsesMode: string(ResponsesSupportModeForceChatCompletions), ExtraKeyResponsesSupported: true}, ResponsesSupportNo},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveResponsesSupport(tc.extra); got != tc.want {
				t.Errorf("ResolveResponsesSupport(%v) = %v, want %v", tc.extra, got, tc.want)
			}
		})
	}
}

func TestShouldUseResponsesAPI(t *testing.T) {
	tests := []struct {
		name  string
		extra map[string]any
		want  bool
	}{
		{"legacy unknown preserves responses", nil, true},
		{"legacy empty preserves responses", map[string]any{}, true},
		{"probe metadata alone does not alter legacy routing", map[string]any{ExtraKeyResponsesProbeStatus: string(ResponsesProbeStatusVerified)}, true},
		{"explicitly supported", map[string]any{ExtraKeyResponsesSupported: true}, true},
		{"explicitly unsupported", map[string]any{ExtraKeyResponsesSupported: false}, false},
		{"verified probe cannot override explicit false", map[string]any{ExtraKeyResponsesSupported: false, ExtraKeyResponsesProbeStatus: string(ResponsesProbeStatusVerified)}, false},
		{"legacy explicit support", map[string]any{ExtraKeyLegacyUseResponsesAPI: true}, true},
		{"force responses overrides explicit false", map[string]any{ExtraKeyResponsesMode: string(ResponsesSupportModeForceResponses), ExtraKeyResponsesSupported: false}, true},
		{"force chat overrides explicit true", map[string]any{ExtraKeyResponsesMode: string(ResponsesSupportModeForceChatCompletions), ExtraKeyResponsesSupported: true}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ShouldUseResponsesAPI(tc.extra); got != tc.want {
				t.Errorf("ShouldUseResponsesAPI(%v) = %v, want %v", tc.extra, got, tc.want)
			}
		})
	}
}

func TestResponsesProbeStatus(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want ResponsesProbeStatus
	}{
		{"empty", "", ResponsesProbeStatusUnknown},
		{"verified", "verified", ResponsesProbeStatusVerified},
		{"unsupported", "unsupported", ResponsesProbeStatusUnsupported},
		{"degraded", "degraded", ResponsesProbeStatusDegraded},
		{"invalid", "healthy", ResponsesProbeStatusUnknown},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeResponsesProbeStatus(tc.raw); got != tc.want {
				t.Errorf("NormalizeResponsesProbeStatus(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}

	extra := map[string]any{ExtraKeyResponsesProbeStatus: string(ResponsesProbeStatusDegraded)}
	if got := ResolveResponsesProbeStatus(extra); got != ResponsesProbeStatusDegraded {
		t.Fatalf("ResolveResponsesProbeStatus() = %q", got)
	}
}

func TestNormalizeResponsesSupportMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want ResponsesSupportMode
	}{
		{"empty", "", ResponsesSupportModeAuto},
		{"auto", "auto", ResponsesSupportModeAuto},
		{"force responses", "force_responses", ResponsesSupportModeForceResponses},
		{"force chat completions", "force_chat_completions", ResponsesSupportModeForceChatCompletions},
		{"invalid", "enabled", ResponsesSupportModeAuto},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeResponsesSupportMode(tc.mode); got != tc.want {
				t.Errorf("NormalizeResponsesSupportMode(%q) = %q, want %q", tc.mode, got, tc.want)
			}
		})
	}
}
