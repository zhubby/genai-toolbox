package resources

import (
	"github.com/googleapis/genai-toolbox/internal/auth"
	"github.com/googleapis/genai-toolbox/internal/prompts"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

// Manager is the runtime resource access interface used by the server.
// Method names match the existing ResourceManager API so call sites remain stable.
//
// Implementations:
// - ResourceManager: in-memory snapshot (legacy)
// - ResourceManagerForDB: reads from config DB on demand
type Manager interface {
	GetSource(sourceName string) (sources.Source, bool)
	GetAuthService(authServiceName string) (auth.AuthService, bool)
	GetTool(toolName string) (tools.Tool, bool)
	GetToolset(toolsetName string) (tools.Toolset, bool)
	GetPrompt(promptName string) (prompts.Prompt, bool)
	GetPromptset(promptsetName string) (prompts.Promptset, bool)

	GetAuthServiceMap() map[string]auth.AuthService
	GetToolsMap() map[string]tools.Tool
	GetPromptsMap() map[string]prompts.Prompt
}


