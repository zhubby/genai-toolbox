package server

import "time"

type configDBSourceDTO struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type configDBToolDTO struct {
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	SourceName *string        `json:"sourceName,omitempty"`
	Config     map[string]any `json:"config"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

type configDBAuthServiceDTO struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type configDBToolsetDTO struct {
	Name      string    `json:"name"`
	ToolNames []string  `json:"toolNames"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type configDBPromptDTO struct {
	Name      string         `json:"name"`
	Kind      string         `json:"kind"`
	Config    map[string]any `json:"config"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type configDBPromptsetDTO struct {
	Name        string    `json:"name"`
	PromptNames []string  `json:"promptNames"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type configDBCreateSourceRequest struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type configDBUpdateSourceRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type configDBCreateToolRequest struct {
	Name       string         `json:"name"`
	Kind       string         `json:"kind"`
	SourceName *string        `json:"sourceName"`
	Config     map[string]any `json:"config"`
}

type configDBUpdateToolRequest struct {
	Kind       string         `json:"kind"`
	SourceName *string        `json:"sourceName"`
	Config     map[string]any `json:"config"`
}

type configDBCreateAuthServiceRequest struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type configDBUpdateAuthServiceRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type configDBCreateToolsetRequest struct {
	Name      string   `json:"name"`
	ToolNames []string `json:"toolNames"`
}

type configDBUpdateToolsetRequest struct {
	ToolNames []string `json:"toolNames"`
}

type configDBCreatePromptRequest struct {
	Name   string         `json:"name"`
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type configDBUpdatePromptRequest struct {
	Kind   string         `json:"kind"`
	Config map[string]any `json:"config"`
}

type configDBCreatePromptsetRequest struct {
	Name        string   `json:"name"`
	PromptNames []string `json:"promptNames"`
}

type configDBUpdatePromptsetRequest struct {
	PromptNames []string `json:"promptNames"`
}

