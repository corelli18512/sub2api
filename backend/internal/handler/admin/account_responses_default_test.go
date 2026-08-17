package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestDefaultOpenAIAPIKeyResponsesCapability(t *testing.T) {
	t.Run("new OpenAI API-key accounts default to Chat Completions", func(t *testing.T) {
		extra := defaultOpenAIAPIKeyResponsesCapability(service.PlatformOpenAI, service.AccountTypeAPIKey, nil)
		require.Equal(t, false, extra[openai_compat.ExtraKeyResponsesSupported])
	})

	t.Run("explicit false is preserved", func(t *testing.T) {
		extra := map[string]any{openai_compat.ExtraKeyResponsesSupported: false, "source": "marketplace"}
		got := defaultOpenAIAPIKeyResponsesCapability(service.PlatformOpenAI, service.AccountTypeAPIKey, extra)
		require.Equal(t, extra, got)
	})

	t.Run("explicit true is preserved", func(t *testing.T) {
		extra := map[string]any{openai_compat.ExtraKeyResponsesSupported: true}
		got := defaultOpenAIAPIKeyResponsesCapability(service.PlatformOpenAI, service.AccountTypeAPIKey, extra)
		require.Equal(t, true, got[openai_compat.ExtraKeyResponsesSupported])
	})

	t.Run("trusted override is preserved", func(t *testing.T) {
		extra := map[string]any{openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceResponses)}
		got := defaultOpenAIAPIKeyResponsesCapability(service.PlatformOpenAI, service.AccountTypeAPIKey, extra)
		require.NotContains(t, got, openai_compat.ExtraKeyResponsesSupported)
	})

	t.Run("other account types are unchanged", func(t *testing.T) {
		require.Nil(t, defaultOpenAIAPIKeyResponsesCapability(service.PlatformOpenAI, service.AccountTypeOAuth, nil))
		require.Nil(t, defaultOpenAIAPIKeyResponsesCapability(service.PlatformAnthropic, service.AccountTypeAPIKey, nil))
	})
}
