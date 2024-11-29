
FROM golang:1.23 as ollamanager-build

WORKDIR /go/src/app
COPY . .

RUN go mod download &&\
  CGO_ENABLED=0 go build -o /go/bin/ollamanager

FROM gcr.io/distroless/static-debian11:nonroot

ENV TERM=xterm-256color
COPY --from=ollamanager-build /go/bin/ollamanager /
ENTRYPOINT ["/ollamanager"]

