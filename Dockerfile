FROM golang

WORKDIR /home/golang



COPY . .

ENV GOPROXY=https://goproxy.cn
ENV GO111MODULE="on"

RUN go build cmd/othello.go


ENTRYPOINT ["./othello"]