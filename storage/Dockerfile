FROM debian:bookworm-slim

WORKDIR /app

# Copy the precompiled binary
COPY bin/storage /app/storage

# Make the binary executable
RUN chmod +x /app/storage

# Run the storage with config file option
ENTRYPOINT ["/app/storage"]
CMD ["-c", "/etc/halko/halko.cfg"]
