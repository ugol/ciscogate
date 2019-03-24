FROM centos:7

USER root

RUN yum -y update && yum clean all

RUN mkdir -p /go && chmod -R 777 /go && \
    yum install -y centos-release-scl && \
    yum -y install git go-toolset-7-golang && yum clean all

ENV GOPATH=/go \
    BASH_ENV=/opt/rh/go-toolset-7/enable \
    ENV=/opt/rh/go-toolset-7/enable \
    PROMPT_COMMAND=". /opt/rh/go-toolset-7/enable"

WORKDIR /go

#FROM centos/go-toolset-7-centos7
#FROM golang:1.12.1-alpine3.9

#WORKDIR /go
#WORKDIR .
COPY . .

RUN go build
RUN go get -u

CMD ["./ciscogate start"]
