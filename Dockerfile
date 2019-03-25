FROM centos
USER root
RUN  yum install -y git && \
     yum clean all

RUN cd /tmp/
    curl -O https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.12.1.linux-amd64.tar.gz 
    export GOPATH=$HOME/work
    export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
    go version

COPY . .

RUN go build &&\
    go get -u

CMD ["./ciscogate start"]
