package utils

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/napuu/gpsp-bot/internal/config"
)

type CutVideoArgs struct {
	StartMinutes    float64 `json:"start_minutes"`
	StartSeconds    float64 `json:"start_seconds"`
	DurationMinutes float64 `json:"duration_minutes"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type ToolCall struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Function json.RawMessage `json:"function"`
}

type Choice struct {
	Index     int        `json:"index"`
	Message   Message    `json:"message"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func ParseCutArgs(msg string) (float64, float64, error) {
	if len(msg) <= 3 {
		return 0, 0, nil
	}

	client := resty.New()

	requestPayload := map[string]any{
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"schema": map[string]any{
					"title": "VideoCutInstruction",
					"type":  "object",
					"properties": map[string]any{
						"start_minutes": map[string]any{
							"title": "Start Minutes",
							"type":  "number",
						},
						"start_seconds": map[string]any{
							"title": "Start Seconds",
							"type":  "number",
						},
						"duration_minutes": map[string]any{
							"title": "Duration Minutes",
							"type":  "number",
						},
						"duration_seconds": map[string]any{
							"title": "Duration Seconds",
							"type":  "number",
						},
					},
					"required":             []string{"start_minutes", "start_seconds"},
					"additionalProperties": false,
				},
				"name":   "videoCutInstruction",
				"strict": true,
			},
		},
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `Cut video with subsecond level accuracy. Instructions are likely in English or Finnish.
				Some examples:
				* 1m33s- => start_minutes = 1, start_seconds = 60
				* 20s-45s- => start_minutes = 0, start_seconds = 20, duration_minutes = 0, duration_seconds = 25
				* vikat 2m34s => start_minutes = -2, start_seconds = -34
				* ekat 6m8s => start_minutes = 0, start_seconds = 0, duration_minutes = 6, duration_seconds = 8
				* 1m3.5s- => start_minutes = 1, start_seconds = 3.5
				* last 15s => start_minutes = 0, start_seconds = -15
				`,
			},
			{"role": "user", "content": msg},
		},
		"model": "mistral-large-latest",
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.FromEnv().MISTRAL_TOKEN)).
		SetBody(requestPayload).
		Post("https://api.mistral.ai/v1/chat/completions")

	if err != nil {
		return 0, 0, err
	}

	var response Response
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return 0, 0, err
	}

	if len(response.Choices) == 0 {
		return 0, 0, fmt.Errorf("no choices in response")
	}

	// Parse the JSON content from the response
	var args CutVideoArgs
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), &args); err != nil {
		return 0, 0, err
	}

	// Calculate start and duration in seconds
	startOnlySeconds := args.StartMinutes*60 + args.StartSeconds
	var durationOnlySeconds float64
	if args.DurationMinutes > 0 || args.DurationSeconds > 0 {
		durationOnlySeconds = args.DurationMinutes*60 + args.DurationSeconds
	}

	return startOnlySeconds, durationOnlySeconds, nil
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

func GetNegation(input string) string {
	client := resty.New()

	requestPayload := map[string]any{
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `Olet botti joka palauttaa virkkeen käänteisellä merkityksellä. Seuraa näitä ohjeita:
- Saat luvan lisätä vastaukseen nimen vain jos se esiintyy myös käyttäjän viimeisessä viestissä.
- Pyri säilyttämään alkuperäinen kirjoitustyyli.
- Jos virkkeessä on useampi lause, palauta kielteinen muoto kaikista niistä.
- Jos lauseessa on nimetty subjekti, käytä kolmatta persoonaa kieltosanalla "ei": esim. "Mikko ei mene".
- Jos lauseessa on ensimmäisen persoonan verbi, käytä kieltosanaa "en": esim. "en mene", "en osta".
- Jos lauseessa ei ole selkeää subjektia tai verbiä, käytä passiivia kieltosanalla "ei": esim. "ei mennä", "ei laiteta".
- Jaan on suomalainen miehen nimi.`,
			},
			{"role": "user", "content": "mikko menee töihin"},
			{"role": "assistant", "content": "mikko ei mene töihin"},
			{"role": "user", "content": "auto ostoon"},
			{"role": "assistant", "content": "ei laiteta autoa ostoon"},
			{"role": "user", "content": "takaisin töihin"},
			{"role": "assistant", "content": "ei mennä takaisin töihin"},
			{"role": "user", "content": "esitän puhelimessa mikko mallikasta ja jätän 200$ tarjouksen"},
			{"role": "assistant", "content": "en esitä puhelimessa mikko mallikasta enkä jätä 200$ tarjousta"},
			{"role": "user", "content": "200k tarjous menemään"},
			{"role": "assistant", "content": "ei laiteta 200k tarjousta menemään"},
			{"role": "user", "content": input},
		},
		"model": "mistral-large-latest",
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.FromEnv().MISTRAL_TOKEN)).
		SetBody(requestPayload).
		Post("https://api.mistral.ai/v1/chat/completions")

	if err != nil {
		log.Fatal(err)
	}

	var response Response
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		log.Fatal(err)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content
	}

	return "Hyvä prompti..."
}
