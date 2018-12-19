FROM golang:latest as builder
WORKDIR /go-modules

ARG SSH_KEY
ARG SSH_KEY_PASSPHRASE

COPY . ./

RUN mkdir -p /root/.ssh && \
    chmod 0700 /root/.ssh && \
    ssh-keyscan bitbucket.org > /root/.ssh/known_hosts && \
    echo "${SSH_KEY}" > /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa && \
    git config --global url."git@bitbucket.org:".insteadOf "https://bitbucket.org/" && \
    GOOS=linux go build -a -installsuffix cgo -ldflags "-linkmode external -extldflags -static" -o app

FROM scratch
EXPOSE 8080
COPY --from=builder /go-modules/app app
ENTRYPOINT ["/app"]

