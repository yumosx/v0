package service

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

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
	//go:embed prompt/human.txt
	askDesc string
	//go:embed prompt/txt.txt
	txt string
)

type Plan struct {
	handler *llm.Handler
}

func NewPlan(handler *llm.Handler) *Plan {
	return &Plan{handler: handler}
}

type Argument struct {
	Text string `json:"text"`
}

func (p *Plan) Execute(ctx context.Context, input string) error {
	var req domain.LLMRequest
	req.Tools = []domain.Tool{p.newAskHuman(), p.newPlanTool()}
	req.Msgs = []domain.Msg{
		{Role: "system", Content: systemContent},
		{Role: "user", Content: input},
	}

	invoke, err := p.handler.Invoke(ctx, req)
	if err != nil {
		return err
	}

	if len(invoke.ToolCalls) == 0 {
		return nil
	}

	var msgs []domain.Msg
	for _, tool := range invoke.ToolCalls {
		switch tool.Function.Name {
		case "question":
			var text Argument
			_ = json.Unmarshal([]byte(tool.Function.Arguments), &text)

			amsg := domain.Msg{Role: "assistant", Content: invoke.Content, ToolCalls: []domain.LLMToolCall{p.AssistantMsg(tool)}}
			msg := p.AskHuman(tool)
			msgs = []domain.Msg{amsg, msg}
			break
		case "plan":
			//如果第一次已经提供了足够的信息, 那么直接返回
			return p.PrintPlan(tool.Function.Arguments)
		default:
			return nil
		}
	}
	// 进入一个不断和大模型交互的循环
	result := p.Loop(ctx, msgs)
	println(result)
	return nil
}

func (p *Plan) Loop(ctx context.Context, msg []domain.Msg) string {
	var req domain.LLMRequest
	req.Tools = []domain.Tool{p.newPlanTool(), p.newAskHuman()}
	req.Msgs = msg
	for {
		invoke, err := p.handler.Invoke(ctx, req)
		if err != nil {
			fmt.Println(err.Error())
			return "调用模型出现错误"
		}
		for _, tool := range invoke.ToolCalls {
			switch tool.Function.Name {
			case "question":
				amsg := domain.Msg{Role: "assistant", Content: invoke.Content, ToolCalls: []domain.LLMToolCall{p.AssistantMsg(tool)}}
				msg := p.AskHuman(tool)
				req.Msgs = []domain.Msg{amsg, msg}
				break
			case "plan":
				err := p.PrintPlan(tool.Function.Arguments)
				if err != nil {
					return "生成计划失败"
				}
				return "已经返回所谓的执行计划"
			}
		}
	}
}

func (p *Plan) input(inquire string) string {
	fmt.Printf("Bot: %s\n\nYou: ", inquire)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	return strings.TrimSpace(response)
}

func (p *Plan) AskHuman(toolCall domain.LLMToolCall) domain.Msg {
	//tool calls msg
	var msg domain.Msg
	msg.Role = "tool"
	msg.Id = toolCall.ID
	msg.Name = toolCall.Function.Name

	var arg Argument
	_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &arg)

	msg.Content = p.input(arg.Text)
	return msg
}

func (p *Plan) AssistantMsg(toolCall domain.LLMToolCall) domain.LLMToolCall {
	//var arg Argument
	//_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &arg)

	return domain.LLMToolCall{
		Index: toolCall.Index,
		ID:    toolCall.ID,
		Type:  "function",
		Function: domain.LLMToolCallFunction{
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		},
	}
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
		Name:        "question",
		Description: askDesc,
		Parameters: &domain.FunctionParameters{
			Type:       "object",
			Properties: p.newAskParams(),
			Required:   []string{"text"},
		},
	}
}

func (p *Plan) newAskParams() *domain.Parameters {
	var params = domain.NewParams()
	params.Params["text"] = domain.NewValue("string", txt)
	return params
}

func (p *Plan) newPlanParams() *domain.Parameters {
	var params = domain.NewParams()

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

func (p *Plan) PrintPlan(args string) error {
	var (
		parsedArgs map[string]interface{}
		err        error
	)

	if err = json.Unmarshal([]byte(args), &parsedArgs); err != nil {
		return err
	}

	cmd := parsedArgs["command"].(string)
	if cmd != "create" {
		return errors.New("args 非 create")
	}

	title := parsedArgs["title"].(string)
	println(title)
	steps := parsedArgs["steps"].([]interface{})
	for _, step := range steps {
		if s, ok := step.(string); ok {
			println(s)
		}
	}
	return nil
}
