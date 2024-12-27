FROM ubuntu:22.04

WORKDIR /asn-compiler-dev

# dependencies
RUN DEBIAN_FRONTEND=noninteractive apt -y update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y build-essential wget git; \
    DEBIAN_FRONTEND=noninteractive apt clean

RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz && \
    tar -C /etc -xzf go1.23.4.linux-amd64.tar.gz && \
    rm -f go1.23.4.linux-amd64.tar.gz
ENV PATH="${PATH}:/etc/go/bin"

# TODO enable this if private repos are used
#RUN git config --global --add url."git@github.com:".insteadOf "https://github.com/"
#ENV GOPRIVATE="github.com/amianetworks/*"
#RUN mkdir -p /root/.ssh
#COPY ./ssh /root/.ssh
#RUN chmod 400 /root/.ssh/id_rsa

# Plugin # TODO change this as needed
RUN mkdir -p controller/build/plugins controller/build/config node/build/plugins servicenode/build/config
RUN cd services/<PLUGIN>; \
    make build; \
    cp build/controller/*.so ../../controller/build/plugins/; \
    cp build/controller/*.conf ../../controller/build/config/; \
    cp build/servicenode/*.so ../../node/build/plugins/; \
    cp build/servicenode/*.conf ../../servicenode/build/config/

ENTRYPOINT ["bash"]
