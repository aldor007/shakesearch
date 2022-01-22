FROM golang:1.15
ADD . /go/src


RUN cd /go/src/ ;go mod download; go build -o /go/shake-search main.go

ENTRYPOINT ["/go/shake-search"]

# Expose the server TCP port
EXPOSE 3001