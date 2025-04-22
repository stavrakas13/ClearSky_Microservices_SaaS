FROM golang:1.24

WORKDIR /grade_review_request_service

COPY . .

RUN go mod init grade_review_request_service  && \
    go mod tidy && \
    go build -o main .

EXPOSE 8081

CMD ["./main"]