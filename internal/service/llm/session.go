package llm

import "github.com/cohesion-org/deepseek-go"

type Session struct {
	messages []deepseek.ChatCompletionMessage
}

func NewSession() *Session {
	return &Session{messages: make([]deepseek.ChatCompletionMessage, 0, 10)}
}

func (session *Session) AppendMsg(msg deepseek.ChatCompletionMessage) {
	session.messages = append(session.messages, msg)
}
