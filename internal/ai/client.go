package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/harish551/editpilot/internal/config"
	"github.com/harish551/editpilot/internal/models"
)

type Planner interface {
	BuildPlan(ctx context.Context, inputs []string, prompt string, output string) (models.Plan, error)
}

type Client struct {
	cfg        config.Config
	httpClient *http.Client
}

func New(cfg config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (c *Client) BuildPlan(ctx context.Context, inputs []string, prompt string, output string) (models.Plan, error) {
	switch strings.ToLower(c.cfg.AIProvider) {
	case "openai":
		return c.buildPlanOpenAI(ctx, inputs, prompt, output)
	default:
		return models.Plan{}, fmt.Errorf("unsupported AI provider: %s", c.cfg.AIProvider)
	}
}

func (c *Client) buildPlanOpenAI(ctx context.Context, inputs []string, prompt string, output string) (models.Plan, error) {
	if strings.TrimSpace(c.cfg.APIKey) == "" {
		return models.Plan{}, fmt.Errorf("EDITPILOT_AI_API_KEY is required for OpenAI provider")
	}

	inputObjects := make([]map[string]string, 0, len(inputs))
	for i, in := range inputs {
		inputObjects = append(inputObjects, map[string]string{
			"id":   fmt.Sprintf("input%d", i+1),
			"path": in,
		})
	}

	schemaPrompt := `You are an FFmpeg planning assistant. Convert the user request into strict JSON only.
Return this shape exactly:
{
  "inputs": [{"id":"input1","path":"clip.mp4"}],
  "operations": [
    {"type":"trim","input":"input1","params":{"start":0,"end":5}},
    {"type":"resize","input":"input1","params":{"width":1080,"height":1920}},
    {"type":"text_overlay","input":"input1","params":{"text":"Title","x":"(w-text_w)/2","y":"80","font_size":48}},
    {"type":"speed","input":"input1","params":{"factor":1.25}},
    {"type":"concat","inputs":["input1","input2"]},
    {"type":"mute","input":"input1"}
  ],
  "output": {"path":"output.mp4"}
}
Only use supported operations: trim, resize, text_overlay, speed, concat, mute, noop.
Do not include markdown fences.`

	payload := map[string]any{
		"model": c.cfg.AIModel,
		"messages": []map[string]string{
			{"role": "system", "content": schemaPrompt},
			{"role": "user", "content": fmt.Sprintf("Inputs: %v\nPrompt: %s\nOutput: %s", inputObjects, prompt, output)},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return models.Plan{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return models.Plan{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.Plan{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return models.Plan{}, fmt.Errorf("OpenAI request failed with status %s", resp.Status)
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return models.Plan{}, err
	}
	if len(response.Choices) == 0 {
		return models.Plan{}, fmt.Errorf("OpenAI response had no choices")
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var plan models.Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return models.Plan{}, fmt.Errorf("decode AI plan JSON: %w", err)
	}
	if plan.Output.Path == "" {
		plan.Output.Path = output
	}
	return plan, nil
}
