FROM golang:1.18-alpine as build

RUN mkdir /missing-container-metrics
WORKDIR /missing-container-metrics
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=main
RUN CGO_ENABLED=0 go build -ldflags "-X main.Version=$VERSION" -o missing-container-metrics . 

FROM scratch
COPY --from=build /missing-container-metrics/missing-container-metrics /missing-container-metrics
EXPOSE 3001
ENTRYPOINT ["/missing-container-metrics"]
