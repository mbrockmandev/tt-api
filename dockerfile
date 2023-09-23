FROM golang:1.21

LABEL base.name="tometracker-api"

WORKDIR /app

COPY . .

RUN go build -o main .

EXPOSE 8080

ENTRYPOINT [ "./main" ]


