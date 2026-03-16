package main

import (
	"context"
	"fmt"
	"os"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func main() {
	key := os.Getenv("VOYAGE_API_KEY")
	fmt.Printf("VOYAGE_API_KEY: %s\n", key)
	client := anthropic.NewClient(option.WithAPIKey(key))
	message, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("What is a quaternion?")),
		},
		Model: anthropic.ModelClaudeSonnet4_5,
	})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%+v\n", message.Content)
}
