FROM golang:1.22

WORKDIR /simulator

RUN go install github.com/cosmtrek/air@latest
RUN air init

# Start the back end and listen to changes in files.
ENTRYPOINT air --build.cmd "go build" --build.bin "./simulator"
