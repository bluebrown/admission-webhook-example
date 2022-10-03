FROM golang:1.19
ENV CGO_ENABLED=0
RUN echo 'nonroot:x:65532:65532:nonroot:/nonexistent:/usr/sbin/nologin' >/passwd
WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -ldflags '-w -s' -a -v -o /go/bin/webhooks ./

FROM scratch
COPY --from=0 /passwd /etc/passwd
COPY --from=0 /go/bin/webhooks /webhooks
ENTRYPOINT ["/webhooks"]
USER 65532
