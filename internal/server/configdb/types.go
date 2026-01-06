package configdb

import "time"

type SourceDTO struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type ToolDTO struct {
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	SourceName *string        `json:"sourceName,omitempty"`
	Config     map[string]any `json:"config"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

type AuthServiceDTO struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type ToolsetDTO struct {
	Name      string    `json:"name"`
	ToolNames []string  `json:"toolNames"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PromptDTO struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type PromptsetDTO struct {
	Name        string    `json:"name"`
	PromptNames []string  `json:"promptNames"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateSourceRequest struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type ValidateSourceRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type UpdateSourceRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type CreateToolRequest struct {
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	SourceName *string        `json:"sourceName"`
	Config     map[string]any `json:"config"`
}

type UpdateToolRequest struct {
	Kind       string         `json:"kind"`
	SourceName *string        `json:"sourceName"`
	Config     map[string]any `json:"config"`
}

type CreateAuthServiceRequest struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type UpdateAuthServiceRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type CreateToolsetRequest struct {
	Name      string   `json:"name"`
	ToolNames []string `json:"toolNames"`
}

type UpdateToolsetRequest struct {
	ToolNames []string `json:"toolNames"`
}

type CreatePromptRequest struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type UpdatePromptRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type CreatePromptsetRequest struct {
	Name        string   `json:"name"`
	PromptNames []string `json:"promptNames"`
}

type UpdatePromptsetRequest struct {
	PromptNames []string `json:"promptNames"`
}
