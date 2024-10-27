package main

import (
	"context"

	"github.com/openai/openai-go"
)

func getNegation(input string) string {
	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("Olet botti joka palauttaa virkkeen kielteisellä merkityksellä. Voit muuttaa sanamuotoja tarpeen mukaan. Saat luvan lisätä vastaukseen nimen vain jos se esiintyy myös käyttäjän viimeisessä viestissä. Nimet ovat todennäköisesti suomalaisia etunimiä. Jos virkkeessä on useampi lause, palauta kielteinen muoto kaikista niistä."),
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
		Model: openai.F(openai.ChatModelGPT4oMini),
	})
	if err != nil {
		panic(err.Error())
	}

	return chatCompletion.Choices[0].Message.Content
}
