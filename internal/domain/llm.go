package domain

type LLMRequest struct {
	Msgs   []Msg
	Tools  []Tool
	Choice string
}

type Msg struct {
	Role      string
	Content   string
	Id        string
	Name      string
	ToolCalls []LLMToolCall
}

type Tool struct {
	Type     string
	Function Function
}

type Function struct {
	Name        string
	Description string
	Parameters  *FunctionParameters
}

type FunctionParameters struct {
	Type       string
	Properties *Parameters
	Required   []string
}

type LLMResponse struct {
	Content   string
	Done      bool
	ToolCalls []LLMToolCall
}

type LLMToolCall struct {
	Index    int
	ID       string
	Type     string
	Function LLMToolCallFunction
}

type LLMToolCallFunction struct {
	Name      string
	Arguments string
}
