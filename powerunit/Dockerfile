FROM golang:1.22

WORKDIR /app/powerunit

RUN go install github.com/air-verse/air@v1.52.3
RUN air init

# Start the back end and listen to changes in files.
ENTRYPOINT air --build.cmd "go build" --build.bin "./powerunit" -- -k /app/.halko.cfg
