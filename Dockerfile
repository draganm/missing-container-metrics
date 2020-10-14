FROM golang:1.15.1-alpine3.12 as build
RUN mkdir /missing-container-metrics
WORKDIR /missing-container-metrics
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build  -o missing-container-metrics .

FROM scratch
COPY --from=build /missing-container-metrics/missing-container-metrics /missing-container-metrics
ENTRYPOINT ["/missing-container-metrics"]
EXPOSE 3001
