package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cohesion-org/deepseek-go"
	"vzero/internal/service"
	"vzero/internal/service/llm"
)

func main() {
	token := os.Getenv("token")
	client := deepseek.NewClient(token)
	session := llm.NewSession()
	handler := llm.NewHandler(client, session)
	newPlan := service.NewPlan(handler)
	err := newPlan.Execute(context.Background(), "帮我制定一个 后端面试 优化案例")
	if err != nil {
		fmt.Println(err.Error())
	}
}
