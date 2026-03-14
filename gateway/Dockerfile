FROM golang:1.25 AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gateway ./main.go

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=build /out/gateway /gateway

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/gateway"]

