package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/napuu/gpsp-bot/internal/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func getClient() *openai.Client {
	return openai.NewClient(option.WithAPIKey(config.FromEnv().OPENAI_TOKEN))
}

func GetNegation(input string) string {
	client := getClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(
				`Olet botti joka palauttaa virkkeen käänteisellä merkityksellä.
Voit muuttaa sanamuotoja tarpeen mukaan.
Saat luvan lisätä vastaukseen nimen vain jos se esiintyy myös käyttäjän viimeisessä viestissä.
Nimet ovat todennäköisesti suomalaisia etunimiä.
Jos virkkeessä on useampi lause, palauta kielteinen muoto kaikista niistä.
Jaan on suomalainen miehen nimi.
`),
			openai.UserMessage("mikko menee töihin"),
			openai.AssistantMessage("mikko ei mene töihin"),
			openai.UserMessage("auto ostoon"),
			openai.AssistantMessage("ei laiteta autoa ostoon"),
			openai.UserMessage("takaisin töihin"),
			openai.AssistantMessage("ei mennä takaisin töihin"),
			openai.UserMessage("esitän puhelimessa mikko mallikasta ja jätän 200$ tarjouksen"),
			openai.AssistantMessage("en esitä puhelimessa mikko mallikasta enkä jätä 200$ tarjousta"),
			openai.UserMessage(input),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		slog.Error(err.Error())
		return "hyvä prompti..."
	}

	return chatCompletion.Choices[0].Message.Content
}

type CutVideoArgs struct {
	StartMinutes    float64 `json:"start_minutes"`
	StartSeconds    float64 `json:"start_seconds"`
	DurationMinutes float64 `json:"duration_minutes"`
	DurationSeconds float64 `json:"duration_seconds"`
}

func ParseCutArgs(msg string) (float64, float64, error) {
	client := getClient()
	if len(msg) <= 3 {
		return 0, 0, nil
	}

	// Define the request parameters
	params := openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4o),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(msg),
		}),
		Tools: openai.F([]openai.ChatCompletionToolParam{
			{
				Type: openai.F(openai.ChatCompletionToolTypeFunction),
				Function: openai.F(openai.FunctionDefinitionParam{
					Name: openai.String("cut_video"),
					Description: openai.String(
						`Cut video with subsecond level accuracy. Instructions are likely in English or Finnish.
						Some examples:
						* 1m33s- => start_minutes = 1, start_seconds = 60 
						* 20s-45s- => start_minutes = 0, start_seconds = 20, duration_minutes = 0, duration_seconds = 25
						* vikat 2m34s => start_minutes = -2, start_seconds = -34
						* ekat 6m8s => start_minutes = 0, start_seconds = 0, duration_minutes = 6, duration_seconds = 8
						* 1m3.5s- => start_minutes = 1, start_seconds = 3.5
						`),
					Parameters: openai.F(openai.FunctionParameters{
						"type": "object",
						"properties": map[string]interface{}{
							"start_seconds": map[string]string{
								"type":        "number",
								"description": "Start seconds of resulting clip. SHOULD BE NEGATIVE if instructed to get e.g. last 15s.",
							},
							"start_minutes": map[string]string{
								"type":        "number",
								"description": "Start minutes of resulting clip. SHOULD BE NEGATIVE if instructed to get e.g. last 15s.",
							},
							"duration_seconds": map[string]string{
								"type":        "number",
								"description": "Duration of the resulting clip in seconds or 0 if the clip should continue until end of the video.",
							},
							"duration_minutes": map[string]string{
								"type":        "number",
								"description": "Duration of the resulting clip in minutes or 0 if the clip should continue until end of the video.",
							},
						},
						"required": []string{"start_seconds", "start_minutes"},
					}),
				}),
			},
		}),
	}

	// Execute the OpenAI request
	ctx := context.Background()
	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return 0, 0, err
	}

	toolCalls := completion.Choices[0].Message.ToolCalls
	if len(toolCalls) == 0 {
		return 0, 0, fmt.Errorf("no function call")
	}

	// Parse the function call arguments
	var args CutVideoArgs
	if err := json.Unmarshal([]byte(toolCalls[0].Function.Arguments), &args); err != nil {
		return 0, 0, err
	}

	// Calculate start and duration in seconds
	startOnlySeconds := args.StartMinutes*60 + args.StartSeconds
	var durationOnlySeconds float64
	if args.DurationMinutes > 0 || args.DurationSeconds > 0 {
		dur := args.DurationMinutes*60 + args.DurationSeconds
		durationOnlySeconds = dur
	}

	return startOnlySeconds, durationOnlySeconds, nil
}
