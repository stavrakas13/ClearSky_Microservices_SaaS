#student_request_review_service

FROM golang:1.24

WORKDIR /student_request_review_service

COPY . .

RUN go mod init student_request_review_service  && \
    go mod tidy && \
    go build -o main .

EXPOSE 8087

CMD ["./main"]