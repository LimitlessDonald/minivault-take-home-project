package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type GenerateRequestBody struct {
	Prompt string `json:"prompt"`
}

type GenerateResponseBody struct {
	Response string `json:"response"`
}

const (
	InvalidRequestBody = "Invalid request body"
	SomethingWentWrong = "Something went wrong"
)

// IsValidPort Checks if a provided port number is a valid port
func IsValidPort(port int) bool {
	if port > 0 && port <= 65535 {
		return true
	}

	return false
}

// RequestLoggerMiddleware is a gin Middleware for logging API requests to file
func RequestLoggerMiddleware(filename string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody GenerateRequestBody
		// Cast json body to struct
		err := c.ShouldBindBodyWith(&requestBody, binding.JSON)
		if err != nil {
			log.Println(err)
			return
		}
		// Log request
		logData := map[string]interface{}{
			"endpoint":  c.Request.URL.Path,
			"timestamp": time.Now().Unix(),
			"prompt":    requestBody.Prompt,
		}

		LogToFile(filename, logData)
		// move on with processing request
		c.Next()

	}
}

func logResponse(endpoint string, filename string, response GenerateResponseBody) {

	logData := map[string]interface{}{
		"endpoint":  endpoint,
		"timestamp": time.Now().Unix(),
		"response":  response.Response,
	}

	LogToFile(filename, logData)
}

// LogToFile logs data to a file
func LogToFile(filename string, data map[string]interface{}) {

	// create filename dir path if it does not exist
	dir := filepath.Dir(filename)

	dirErr := os.MkdirAll(dir, 0750)
	if dirErr != nil {
		log.Print(dirErr)
	}
	// Open specified file name
	// Create it if it doesn't already exist, open the file for writing only, append to files instead of overwriting it
	// Set permission so only authorized user can gain access to the log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0640)
	if err != nil {
		log.Println(err)
		return
	}

	// Ensures files is closed when done with it to prevent resource leaks, like unnecessary memory /cpu usage
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			log.Println(closeErr)
		}
	}(file)

	// convert data to valid JSON string
	jsonData, jsonErr := json.Marshal(data)
	if jsonErr != nil {
		log.Println(err)
		return
	}

	// Write to file
	_, err = file.WriteString(string(jsonData) + "\n")
	if err != nil {
		log.Println(err)
		return
	}

}

// Handles /generate requests
func generateHandler(c *gin.Context) {
	var requestBody GenerateRequestBody
	// ensure body is available after passing through middleware that's why I am using c.ShouldBindBodyWith
	err := c.ShouldBindBodyWith(&requestBody, binding.JSON)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": InvalidRequestBody})
		return
	}

	response, responseErr := GetLlmResponse(requestBody.Prompt)
	if responseErr != nil {
		log.Println(responseErr)
		// Don't output LLM error directly to API response, as that could pose a potential security risk or reveal sensitive information
		// Output generic message
		c.AbortWithStatusJSON(http.StatusBadRequest, SomethingWentWrong)
	}
	logResponse(c.Request.URL.Path, logFileFlagVal, GenerateResponseBody{Response: response})
	c.AbortWithStatusJSON(http.StatusOK, GenerateResponseBody{Response: response})
	return

}

func GetLlmResponse(prompt string) (string, error) {
	// if stubbed=true in CLI , no need to go ahead with making request to LLM
	if stubbedFlagVal == true {
		return "I'm a local AI model, running offline!", nil
	}
	client := openai.NewClient(
		option.WithBaseURL(llmAPIBaseUrlFlagVal),
	)
	// Using a placeholder context to speed up development
	// Request below is the same as:
	//client.chat.completions.create(
	//    messages=[
	//        {
	//            'role': 'user',
	//            'content': prompt,
	//        }
	//    ],
	//    model=llmModelFlagVal,
	//)
	response, responseErr := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Model: llmModelFlagVal,
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(prompt),
					},
				},
			},
		},
	})

	if responseErr != nil {
		return "", responseErr
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

var (
	portFlagVal          int
	llmAPIBaseUrlFlagVal string
	llmModelFlagVal      string
	logFileFlagVal       string
	testPromptFlagVal    string
	stubbedFlagVal       bool
)

func init() {

	// define CLI flags
	flag.IntVar(&portFlagVal, "port", 8080, "Port to run the server on")
	flag.StringVar(&llmAPIBaseUrlFlagVal, "llm-base-url", "http://localhost:11434/v1/", "Base url of LLM server like Ollama")
	flag.StringVar(&llmModelFlagVal, "llm-model", "llama3.2:latest", "The LLM model to use")
	flag.StringVar(&logFileFlagVal, "log-file", "./logs/log.jsonl", "Log file for requests and responses")
	flag.StringVar(&testPromptFlagVal, "test-prompt", "", "Used to send a test prompt and get a response")
	flag.BoolVar(&stubbedFlagVal, "stubbed", false, "Boolean value to tell the server if it should show stubbed response, or send request to LLM server (default false)")

}
func main() {

	// Parse and prepare the flags for use
	flag.Parse()

	if IsValidPort(portFlagVal) == false {
		log.Fatalf("Invalid port number:%d, Port number must be in between 1 and 65535 ", portFlagVal)
	}

	// Check if url is valid
	_, llmAPIBaseUrlErr := url.Parse(llmAPIBaseUrlFlagVal)
	if llmAPIBaseUrlErr != nil {
		log.Fatalf("Invalid base url %s", llmAPIBaseUrlFlagVal)
	}

	// if test prompt is not empty , output llm response
	if len(testPromptFlagVal) > 0 {
		testPromptResponse, testPromptResponseErr := GetLlmResponse(testPromptFlagVal)
		if testPromptResponseErr != nil {
			fmt.Println(testPromptResponseErr)
		}
		fmt.Println(testPromptResponse)
	}

	// Only run server if we are not running test prompt
	if len(testPromptFlagVal) == 0 {
		router := gin.New()

		router.Use(RequestLoggerMiddleware(logFileFlagVal))

		// Add generate endpoint
		router.POST("/generate", generateHandler)

		router.Run(fmt.Sprintf(":%d", portFlagVal))

	}

}

