# syntax=docker/dockerfile:1

FROM golang:1.20

# Set destination for COPY
WORKDIR /home/micro/sambashare/project/backend/microservice/db

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Build
RUN GO111MODULE=on CGO_ENABLED=1 GOOS=linux go build -o /microservicedb



RUN apt-get update && apt-get install -y --no-install-recommends \
	unzip \
	libaio1 \
	&& apt-get clean \
        && rm -rf /var/lib/apt/lists/* \
        && wget https://github.com/thbono/go-oracle/raw/master/instantclient-basic-linux.x64-11.2.0.4.0.zip \
	&& wget https://github.com/thbono/go-oracle/raw/master/instantclient-sdk-linux.x64-11.2.0.4.0.zip \
	&& unzip instantclient-basic-linux.x64-*.zip -d / \
    	&& unzip instantclient-sdk-linux.x64-*.zip -d / \
	&& rm instantclient-*-linux.x64-*.zip \
    	&& ln -s /instantclient_11_2/libclntsh.so.11.1 /instantclient_11_2/libclntsh.so


ENV LD_LIBRARY_PATH /instantclient_11_2


RUN go mod tidy

RUN go get -d -v gopkg.in/rana/ora.v4



# Optional:
# To bind to a TCP port, runtime parameters must be supplied to the docker command.
# But we can document in the Dockerfile what ports
# the application is going to listen on by default.
# https://docs.docker.com/engine/reference/builder/#expose
#EXPOSE 8080

# Run
CMD ["/microservicedb"]
