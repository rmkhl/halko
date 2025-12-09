FROM debian:bookworm-slim

WORKDIR /app

# Copy the precompiled binary
COPY bin/powerunit /app/powerunit

# Make the binary executable
RUN chmod +x /app/powerunit

# Run the powerunit with config file option
ENTRYPOINT ["/app/powerunit"]
CMD ["-c", "/etc/halko/halko.cfg"]
