FROM golang:1.23.5

WORKDIR /

COPY . .

RUN apt-get update && apt-get install -y git curl

RUN go build -o minivault-server .


RUN mv minivault  /usr/local/bin/



RUN curl -fsSL https://ollama.com/install.sh | sh


CMD ["/bin/bash","-c","ollama run llama3.2:latest & minivault-server"]
