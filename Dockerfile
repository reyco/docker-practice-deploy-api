FROM golang:1.4

RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson

RUN go get github.com/emicklei/go-restful
RUN go get github.com/emicklei/go-restful/swagger

RUN go get github.com/ant0ine/go-json-rest/rest
RUN go get github.com/dgrijalva/jwt-go

# Copy the local package files to the container's workspace.
ADD . /go/src/api

WORKDIR /go/src/api

EXPOSE 8080

ENTRYPOINT ["go", "run", "/go/src/api/main.go"]


