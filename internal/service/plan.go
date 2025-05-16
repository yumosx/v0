package service

import (
	"context"
	_ "embed"

	"vzero/internal/domain"
	"vzero/internal/service/llm"
)

var (
	//go:embed prompt/desc.txt
	description string
	//go:embed prompt/system.txt
	systemContent string
	//go:embed prompt/plan_id.txt
	planId string
)

type Plan struct {
	handler llm.Handler
}

func NewPlan(handler llm.Handler) *Plan {
	return &Plan{handler: handler}
}

func (p *Plan) Execute(ctx context.Context) error {
	var req domain.LLMRequest
	req.Tools = []domain.Tool{p.newPlanTool(), p.newAskHuman()}
	invoke, err := p.handler.Invoke(ctx, req)
	if err != nil {
		return err
	}

	for _, tool := range invoke.ToolCalls {
		switch tool.Function.Name {
		case "ask_human":
		case "plan":
		}
	}

	return nil
}

func (p *Plan) newPlanTool() domain.Tool {
	return domain.Tool{
		Type:     "function",
		Function: p.newPlanFunction(),
	}
}

func (p *Plan) newAskHuman() domain.Tool {
	return domain.Tool{
		Type:     "function",
		Function: p.newAshHuman(),
	}
}

func (p *Plan) newPlanFunction() domain.Function {
	return domain.Function{
		Name:        "plan",
		Description: description,
		Parameters: &domain.FunctionParameters{
			Type:       "object",
			Properties: p.newPlanParams(),
			Required:   []string{"command"},
		},
	}
}

func (p *Plan) newAshHuman() domain.Function {
	return domain.Function{
		Name:        "ask_human",
		Description: "use this tool to ask human for help",
		Parameters: &domain.FunctionParameters{
			Type:       "object",
			Properties: p.newHumanParams(),
			Required:   []string{"inquire"},
		},
	}
}

func (p *Plan) newPlanParams() *domain.Parameters {
	var params *domain.Parameters
	params.Params["plan_id"] = domain.NewValue("string", "unique identifier for t")
	return params
}

func (p *Plan) newHumanParams() *domain.Parameters {
	var params *domain.Parameters

	params.Params["command"] = domain.NewValue("string",
		"The command to execute. Available commands: create, update, list, get, set_active, mark_step, delete.",
		domain.WithEnum([]string{
			"create",
			"update",
			"list",
			"get",
			"set_active",
			"mark_step",
			"delete",
		}),
	)
	params.Params["plan_id"] = domain.NewValue("string", planId)
	params.Params["title"] = domain.NewValue("string",
		"Title for the plan. Required for create command, optional for update command.")
	params.Params["steps"] = domain.NewValue("array",
		"List of plan steps. Required for create command, optional for update command.",
		domain.WithItem(map[string]string{
			"type": "string",
		}))
	params.Params["step_index"] = domain.NewValue("integer",
		"Index of the step to unique (0-based). Required for mark_step command.")

	return params
}
