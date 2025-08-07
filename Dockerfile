FROM golang:1.23.5

WORKDIR /

COPY . .

RUN apt-get update && apt-get install -y git curl

RUN go build -o minivault-server .


RUN mv minivault-server  /usr/local/bin/



RUN curl -fsSL https://ollama.com/install.sh | sh


CMD ["/bin/bash","-c","ollama serve & sleep 5 && ollama pull llama3.2:latest && minivault-server"]
