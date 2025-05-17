package domain

type Parameters struct {
	Params map[string]*Value
}

func NewParams() *Parameters {
	return &Parameters{Params: make(map[string]*Value)}
}

type Value struct {
	Type string
	Desc string
	Enum []string
	Item map[string]string
}

type ValueOption interface {
	Option(sp *Value)
}

type ValueOptionFunc func(sp *Value)

func (fn ValueOptionFunc) Option(sp *Value) {
	fn(sp)
}

func WithEnum(enum []string) ValueOption {
	return ValueOptionFunc(func(sp *Value) {
		sp.Enum = enum
	})
}

func WithItem(item map[string]string) ValueOption {
	return ValueOptionFunc(func(sp *Value) {
		sp.Item = item
	})
}

func NewValue(ty string, desc string, opts ...ValueOption) *Value {
	sp := &Value{Type: ty, Desc: desc}

	for _, opt := range opts {
		opt.Option(sp)
	}

	return sp
}

func (sp *Value) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["type"] = sp.Type
	if sp.Desc != "" {
		result["description"] = sp.Desc
	}
	if len(sp.Enum) != 0 {
		result["enum"] = sp.Enum
	}
	if len(sp.Item) != 0 {
		result["items"] = sp.Item
	}
	return result
}

func (sp *Parameters) ToMap() map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range sp.Params {
		result[k] = v.ToMap()
	}

	return result
}
