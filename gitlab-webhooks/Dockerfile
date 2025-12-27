FROM golang:1.16 as compiler

WORKDIR "/app"
# copy go mod and sum files
COPY go.mod go.sum ./
# download dependencies
RUN go mod download
# copy source from current directory
COPY . .
# build 
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go


FROM alpine:latest as runtime

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=compiler /app/main .

# copy default config
COPY cmd/config.yml .

# run root executable
CMD ["./main"]