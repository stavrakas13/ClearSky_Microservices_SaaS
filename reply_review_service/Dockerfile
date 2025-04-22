FROM golang:1.24

WORKDIR /review_reply_service

COPY . .

RUN go mod init review_reply_service  && \
    go mod tidy && \
    go build -o main .

EXPOSE 8080

CMD ["./main"]