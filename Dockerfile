FROM golang
RUN echo 'nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin' >/passwd
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o webhook ./

FROM scratch
COPY --from=0 /passwd /etc/passwd
COPY --from=0 /workspace/webhook /webhook
ENTRYPOINT ["/webhook"]
USER nobody
