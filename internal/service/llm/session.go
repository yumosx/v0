package llm

import "github.com/cohesion-org/deepseek-go"

type Session struct {
	messages []deepseek.ChatCompletionMessage
}

func NewSession() *Session {
	return &Session{}
}

func (session *Session) AppendMsg(msg deepseek.ChatCompletionMessage) {
	session.messages = append(session.messages, msg)
}
