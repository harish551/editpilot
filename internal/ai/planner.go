package ai

import (
	"context"

	"github.com/harish551/editpilot/internal/config"
	"github.com/harish551/editpilot/internal/models"
	baseplanner "github.com/harish551/editpilot/internal/planner"
)

type Service struct {
	cfg    config.Config
	client *Client
}

func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:    cfg,
		client: New(cfg),
	}
}

func (s *Service) BuildPlan(ctx context.Context, inputs []string, prompt string, output string) (models.Plan, error) {
	if err := s.cfg.ValidateForAI(); err == nil && s.cfg.APIKey != "" {
		if plan, err := s.client.BuildPlan(ctx, inputs, prompt, output); err == nil {
			return plan, nil
		}
	}
	return baseplanner.BuildPromptPlan(inputs, prompt, output)
}
