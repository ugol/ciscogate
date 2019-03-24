FROM centos

USER root
RUN /bin/bash
RUN /bin/yum -y update && yum clean all

RUN mkdir -p /go && chmod -R 777 /go && \
    yum install -y centos-release-scl && \
    yum -y install git go-toolset-7-golang && yum clean all

ENV GOPATH=/go \
    BASH_ENV=/opt/rh/go-toolset-7/enable \
    ENV=/opt/rh/go-toolset-7/enable \
    PROMPT_COMMAND=". /opt/rh/go-toolset-7/enable"

WORKDIR /go


COPY . .
RUN export GOPATH=$GOPATH:/go

RUN go get
#RUN /opt/rh/go-toolset-7/root/usr/bin/go build
#RUN /opt/rh/go-toolset-7/root/usr/bin/go get -u

CMD ["./ciscogate start"]
