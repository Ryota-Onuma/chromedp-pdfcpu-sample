FROM golang:1.22.0

RUN useradd -m -s /bin/bash app && \
    apt update && \
    mv /usr/local/go /home/app/go && \
    apt install -y chromium && \
    mkdir -p /home/app/bin && \
    mv /usr/bin/chromium /home/app/bin/chromium && \
    apt install fonts-noto-cjk -y

ENV GOPATH=/home/app/go
ENV GOBIN=$GOPATH/bin
ENV BIN=/home/app/bin
ENV PATH=$PATH:$GOBIN
ENV PATH=$PATH:$BIN

RUN chown -R -v app:app /home/app

USER app

WORKDIR /home/app/workspace

COPY --chown=app:app ./go.mod ./go.sum ./

RUN go mod download

COPY --chown=app:app . .

CMD ["go", "run", "main.go"]
