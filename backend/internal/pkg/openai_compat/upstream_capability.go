// Package openai_compat provides endpoint-specific capability policy for
// OpenAI-compatible upstream accounts.
//
// Third-party OpenAI-compatible providers commonly support
// /v1/chat/completions without implementing /v1/responses. Capability probes
// are therefore advisory only: they report observed behavior, but never
// override an explicit administrator configuration or automatically upgrade a
// Chat Completions account to Responses routing.
package openai_compat

// AccountResponsesSupport describes the effective administrator-controlled
// Responses API capability of an OpenAI API-key account.
type AccountResponsesSupport int

const (
	// ResponsesSupportUnknown means no explicit routing decision was configured.
	ResponsesSupportUnknown AccountResponsesSupport = iota
	// ResponsesSupportYes means Responses routing was explicitly enabled.
	ResponsesSupportYes
	// ResponsesSupportNo means Responses routing was explicitly disabled.
	ResponsesSupportNo
)

// ResponsesSupportMode describes an explicit account-level routing override.
type ResponsesSupportMode string

const (
	ResponsesSupportModeAuto                 ResponsesSupportMode = "auto"
	ResponsesSupportModeForceResponses       ResponsesSupportMode = "force_responses"
	ResponsesSupportModeForceChatCompletions ResponsesSupportMode = "force_chat_completions"
)

// ResponsesProbeStatus describes the latest observed /v1/responses probe
// result. It is diagnostic metadata and does not change routing by itself.
type ResponsesProbeStatus string

const (
	ResponsesProbeStatusUnknown     ResponsesProbeStatus = "unknown"
	ResponsesProbeStatusVerified    ResponsesProbeStatus = "verified"
	ResponsesProbeStatusUnsupported ResponsesProbeStatus = "unsupported"
	ResponsesProbeStatusDegraded    ResponsesProbeStatus = "degraded"
)

// Account extra keys.
const (
	ExtraKeyResponsesMode            = "openai_responses_mode"
	ExtraKeyResponsesSupported       = "openai_responses_supported"
	ExtraKeyLegacyUseResponsesAPI    = "use_responses_api"
	ExtraKeyResponsesProbeStatus     = "openai_responses_probe_status"
	ExtraKeyResponsesProbeHTTPStatus = "openai_responses_probe_http_status"
	ExtraKeyResponsesProbeCheckedAt  = "openai_responses_probe_checked_at"
)

// NormalizeResponsesSupportMode normalizes an account-level routing override.
func NormalizeResponsesSupportMode(mode string) ResponsesSupportMode {
	switch ResponsesSupportMode(mode) {
	case ResponsesSupportModeForceResponses:
		return ResponsesSupportModeForceResponses
	case ResponsesSupportModeForceChatCompletions:
		return ResponsesSupportModeForceChatCompletions
	default:
		return ResponsesSupportModeAuto
	}
}

// NormalizeResponsesProbeStatus normalizes diagnostic probe metadata.
func NormalizeResponsesProbeStatus(status string) ResponsesProbeStatus {
	switch ResponsesProbeStatus(status) {
	case ResponsesProbeStatusVerified:
		return ResponsesProbeStatusVerified
	case ResponsesProbeStatusUnsupported:
		return ResponsesProbeStatusUnsupported
	case ResponsesProbeStatusDegraded:
		return ResponsesProbeStatusDegraded
	default:
		return ResponsesProbeStatusUnknown
	}
}

// ResolveResponsesProbeStatus returns the latest diagnostic probe state.
func ResolveResponsesProbeStatus(extra map[string]any) ResponsesProbeStatus {
	if extra == nil {
		return ResponsesProbeStatusUnknown
	}
	status, _ := extra[ExtraKeyResponsesProbeStatus].(string)
	return NormalizeResponsesProbeStatus(status)
}

// ResolveResponsesSupport resolves only trusted, explicit configuration.
// Probe metadata is deliberately ignored so a successful one-off probe cannot
// overwrite an administrator's false setting or silently upgrade routing.
func ResolveResponsesSupport(extra map[string]any) AccountResponsesSupport {
	if extra == nil {
		return ResponsesSupportUnknown
	}
	if mode, ok := extra[ExtraKeyResponsesMode].(string); ok {
		switch NormalizeResponsesSupportMode(mode) {
		case ResponsesSupportModeForceResponses:
			return ResponsesSupportYes
		case ResponsesSupportModeForceChatCompletions:
			return ResponsesSupportNo
		}
	}
	value, ok := extra[ExtraKeyResponsesSupported]
	if !ok {
		value, ok = extra[ExtraKeyLegacyUseResponsesAPI]
	}
	if !ok {
		return ResponsesSupportUnknown
	}
	supported, ok := value.(bool)
	if !ok {
		return ResponsesSupportUnknown
	}
	if supported {
		return ResponsesSupportYes
	}
	return ResponsesSupportNo
}

// ShouldUseResponsesAPI controls native /v1/responses routing. Existing
// accounts without an explicit setting retain the historical Responses
// behavior for compatibility. New API-key accounts are created with an
// explicit false default by the admin handler, so probes cannot opt them in.
func ShouldUseResponsesAPI(extra map[string]any) bool {
	return ResolveResponsesSupport(extra) != ResponsesSupportNo
}
