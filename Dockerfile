FROM centos/go-toolset-7-centos7
#FROM golang:1.12.1-alpine3.9

#WORKDIR /go
#WORKDIR .
COPY . .

RUN go build
RUN go get -u

CMD ["./ciscogate start"]
