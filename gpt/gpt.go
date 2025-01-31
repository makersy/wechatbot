package gpt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/otiai10/openaigo"
	"github.com/qingconglaixueit/wechatbot/config"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const BASEURL = "https://open.haoliny.top/v1/"

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChoiceItem struct {
	Message      openaigo.ChatMessage `json:"message"`
	Text         string               `json:"text"`
	Index        int                  `json:"index"`
	Logprobs     int                  `json:"logprobs"`
	FinishReason string               `json:"finish_reason"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model            string                 `json:"model,omitempty"`
	Messages         []openaigo.ChatMessage `json:"messages,omitempty"`
	Prompt           string                 `json:"prompt,omitempty"`
	MaxTokens        uint                   `json:"max_tokens,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             int                    `json:"top_p,omitempty"`
	FrequencyPenalty int                    `json:"frequency_penalty,omitempty"`
	PresencePenalty  int                    `json:"presence_penalty,omitempty"`
}

// Completions GPT文本模型回复
//curl https://api.openai.com/v1/completions
//-H "Content-Type: application/json"
//-H "Authorization: Bearer your chatGPT key"
//-d '{"model": "text-davinci-003", "prompt": "give me good song", "temperature": 0, "max_tokens": 7}'
func Completions(msg string) (string, error) {
	cfg := config.LoadConfig()
	requestBody := ChatGPTRequestBody{
		Model:            cfg.Model,
		MaxTokens:        cfg.MaxTokens,
		Temperature:      cfg.Temperature,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	var url string
	switch {
	case strings.Contains(requestBody.Model, "gpt-3.5"):
		requestBody.Messages = []openaigo.ChatMessage{{Role: "user", Content: msg}}
		url = BASEURL + "chat/completions"
		break
	default:
		requestBody.Prompt = msg
		url = BASEURL + "completions"
		break
	}

	requestData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	logger.Info(fmt.Sprintf("request gpt json string : %v", string(requestData)))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))

	if err != nil {
		return "", err
	}

	apiKey := config.LoadConfig().ApiKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		return "", errors.New(fmt.Sprintf("请求GPT出错了，gpt api status code not equals 200,code is %d ,details:  %v ", response.StatusCode, string(body)))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	logger.Info(fmt.Sprintf("response gpt json string : %v", string(body)))

	gptResponseBody := &ChatGPTResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		switch {
		case strings.Contains(requestBody.Model, "gpt-3.5"):
			reply = gptResponseBody.Choices[0].Message.Content
		default:
			reply = gptResponseBody.Choices[0].Text
		}
	}
	logger.Info(fmt.Sprintf("gpt response text: %s ", reply))
	return reply, nil
}
