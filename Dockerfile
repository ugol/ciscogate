FROM centos
USER root
RUN  yum install -y git && \
     yum clean all

RUN cd /tmp/ && \
    curl -O https://dl.google.com/go/go1.12.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.12.1.linux-amd64.tar.gz && \
    echo "export GOPATH=$HOME/work" >> /etc/profile &&\
    echo "export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin" >> /etc/profile

COPY . .

RUN source /etc/profile &&\
    go build &&\
    go get -u

RUN mv ciscogate /usr/local/bin/
RUN chgrp -R 0 /usr/local/bin/ciscogate && \
    chmod -R g=u /usr/local/bin/ciscogate

EXPOSE 8080 

CMD /usr/local/bin/ciscogate start
