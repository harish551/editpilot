package models

type InputAsset struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

type Operation struct {
	Type   string         `json:"type"`
	Input  string         `json:"input,omitempty"`
	Inputs []string       `json:"inputs,omitempty"`
	Params map[string]any `json:"params,omitempty"`
}

type OutputSpec struct {
	Path string `json:"path"`
}

type Plan struct {
	Inputs     []InputAsset `json:"inputs"`
	Operations []Operation  `json:"operations"`
	Output     OutputSpec   `json:"output"`
}
