## Improvements To Make The Project Better
- Dockerfile build should be multistage so build doesn't last longer than necessary, it would also reduce final image size 
- Explore the possibility of optimizing further using alpine as base image
- Add short options for terminal flags e.g --port should have the -p option
- Move response logging to a middleware, so response is automatically logged without having to use logResponse in each endpoint handler
- Include API key for llms for security 
- Add API key flag to CLI
- Add versioning / version management when working with docker , either with .env file and/or environment variables 
- Add option to get CLI flags from environment variables when running with docker 
- Add sample docker-compose.yml file 
- Llama.cpp would be faster in my experience (during builds and API requests), wanted to stick to the requirements of the project that's why I used Ollama 
- Write automated go tests and benchmarks to prevent issues in production and performance optimization (Intentionally excluded automated tests and benchmarks because of time constraints, since this is a demo project, I wouldn't do that in production)
- Add /status endpoint 
- Improve modularity 
- Reduce unnecessary repetitions 
- Improve this README
- Add option to indicate preferred model from API request to /generate endpoint
- If model doesn't exist, it just outputs error, if a new model needs to be used it has to be downloaded manually via  `ollama pull`
- Make provision for streaming response


## Building & Installation
Clone this repo : 
```shell
git clone https://github.com/limitlessdonald/minivault-take-home-project && cd minivault-take-home-project
```

 With docker 
```shell
docker build -t minivault:latest .
```
Without Docker
- Install Go if you don't have it already installed :  https://go.dev/doc/install
- Build the go project 
```shell
go build -o minivault-server .
```
- Install Ollama if you don't have it installed, or use any `OpenAI` compatible local LLM server you have running 
- Pull the default model, the project is set to work with , or choose your preferred pre-downloaded model later 
```shell
ollama pull llama3.2:latest
```

## Usage 

### CLI Options
```shell
./minivault-server -h
Usage of ./minivault-server:
  -llm-base-url string
        Base url of LLM server like Ollama (default "http://localhost:11434/v1/")
  -llm-model string
        The LLM model to use (default "llama3.2:latest")
  -log-file string
        Log file for requests and responses (default "./logs/log.jsonl")
  -port int
        Port to run the server on (default 8080)
  -stubbed
        Boolean value to tell the server if it should show stubbed response, or send request to LLM server (default false)
  -test-prompt string
        Used to send a test prompt and get a response

```

### Run With docker : 

- With default parameters 

```shell
    docker run --name minivault-container -p 8080:8080 minivault:latest
```
- With custom CLI parameters 

```shell
       docker run --name minivault-container -p 8080:7272 minivault:latest /bin/bash -c "ollama serve & sleep 5 && ollama pull llama3.2:latest && minivault-server --port=7272 --test-prompt='Tell me about Friedrich Nietzsche'"
```

### Run Without Docker

- With default parameters 

```shell
./minivault-server 
```

```shell
curl -X POST -H "Content-Type: application/json" -d '{"prompt": "What are the top 10 programming languages"}' http://localhost:8080/generate

```

- With custom CLI parameters

```shell
./minivault-server --port=7272 --llm-model="gemma3n:latest"
```

## Design Thoughts 
- Followed project requirements as accurately as possible
- Used constants for demo API responses , to demonstrate the importance of "centralizing" response text, this helps uniformity in API design, and would also aid in translation to other languages if necessary
- Added an option to CLI get a quick stubbed response without making request to LLM server 
- Added `-llm-base-url` option to CLI, so it can work with any `OpenAI` compatible LLM server like `Llama.cpp`, `Ollama` e.t.c 
- The `OpenAI` compatible server approach was chosen to make it easy for "hot swapping" different types of LLM server like mentioned above 
- The `OpenAI` go library was selected to save time and make it relatively easier to make new types of requests, e.g Image/video inference with multimodal LLMs
- If test prompt is used in CLI , `minivault` server does not run
