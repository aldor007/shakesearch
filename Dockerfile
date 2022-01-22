FROM golang:1.15
ADD . /go/src


RUN cd /go/src/ ;go mod download; go build -o shake-search main.go

ENTRYPOINT ["/go/src/shake-search"]

# Expose the server TCP port
EXPOSE 3001