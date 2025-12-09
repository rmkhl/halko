FROM debian:bookworm-slim

WORKDIR /app

# Copy the precompiled binary
COPY bin/executor /app/executor

# Make the binary executable
RUN chmod +x /app/executor

# Run the executor with config file option
ENTRYPOINT ["/app/executor"]
CMD ["-c", "/etc/halko/halko.cfg"]
