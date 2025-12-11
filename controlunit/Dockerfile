FROM debian:bookworm-slim

WORKDIR /app

# Copy the precompiled binary
COPY bin/controlunit /app/controlunit

# Make the binary executable
RUN chmod +x /app/controlunit

# Run the controlunit with config file option
ENTRYPOINT ["/app/controlunit"]
CMD ["-c", "/etc/halko/halko.cfg"]
