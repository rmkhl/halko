FROM debian:bookworm-slim

# Create halko user and group with configurable UID/GID
ARG UID=1000
ARG GID=1000
RUN groupadd -g ${GID} halko && \
    useradd -u ${UID} -g halko -m -s /bin/bash halko

WORKDIR /app

# Copy the precompiled binary
COPY bin/simulator /app/simulator

# Make the binary executable
RUN chmod +x /app/simulator

# Run the simulator with config file options
ENTRYPOINT ["/app/simulator"]
CMD ["-c", "/etc/halko/halko.cfg", "-s", "/etc/halko/simulator.conf"]
