FROM debian:bookworm-slim

WORKDIR /app

# Copy the precompiled binary
COPY bin/simulator /app/simulator

# Make the binary executable
RUN chmod +x /app/simulator

# Run the simulator with config file option
ENTRYPOINT ["/app/simulator"]
CMD ["-c", "/etc/halko/halko.cfg"]
