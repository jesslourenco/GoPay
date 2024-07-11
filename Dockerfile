FROM golang

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o api cmd/gopay/main.go  

EXPOSE 8080

CMD [ "./api" ]