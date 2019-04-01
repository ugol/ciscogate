FROM centos
USER root

# Install go 
RUN curl -O https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.12.1.linux-amd64.tar.gz &&\
    rm -f /go1.12.1.linux-amd64.tar.gz &&\
    echo "export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin" >> /etc/profile &&\
    source /etc/profile

# Compile & configure ciscogate
WORKDIR /tmp/ciscogate
COPY . .
RUN  source /etc/profile &&\
     export GOPATH=$(echo $(pwd)/src) &&\
     go build &&\
     mv ciscogate /usr/local/bin/ &&\
     rm -rf /tmp/ciscogate

RUN chgrp -R 0 /usr/local/bin/ciscogate && \
    chmod -R g=u /usr/local/bin/ciscogate

EXPOSE 8080 

CMD /usr/local/bin/ciscogate start
