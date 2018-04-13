FROM iron/go:dev

ARG BUILD_DATE
ARG VERSION
ARG SERVICEPORT

EXPOSE ${SERVICEPORT}

WORKDIR /app

ENV BUILD_DATE ${BUILD_DATE}
ENV VERSION ${VERSION}
#ENV SRC_DIR=/go/src/github.com/OlegGorj/go-templates-collection/tree/master/rest-api-to-cassandra
#ENV PATH $SRC_DIR:$PATH

ADD . /app
RUN go get github.com/gocql/gocql
RUN go build -o restapiservice service.go
#; cp restapiservice /app/; cp config.json /app/

ENTRYPOINT ["/app/restapiservice"]
