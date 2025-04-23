#instructor_review_reply_service

FROM golang:1.24

WORKDIR /instructor_review_reply_service

COPY . .

RUN go mod init instructor_review_reply_service  && \
    go mod tidy && \
    go build -o main .

EXPOSE 8088

CMD ["./main"]