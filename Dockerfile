FROM debian:buster
RUN apt update && \
    apt install -y curl
# Work inside the /tmp directory
WORKDIR /tmp
RUN curl https://storage.googleapis.com/golang/go1.17.1.linux-amd64.tar.gz -o go.tar.gz && \
    tar -zxf go.tar.gz && \
    rm -rf go.tar.gz && \
    mv go /go
ENV GOPATH /go
ENV PATH $PATH:/go/bin:$GOPATH/bin
# If you enable this, then gcc is needed to debug your app
ENV CGO_ENABLED 1
# TODO: Add other dependencies and stuff here