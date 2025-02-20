# Use an official Java runtime as the base image
FROM openjdk:11-jdk-slim

# Set environment variables for Flink
ENV FLINK_VERSION=1.20 \
    SCALA_VERSION=3.6.2 \
    FLINK_HOME=/opt/flink

# Update package manager and install necessary tools
RUN apt-get update && apt-get install -y \
    curl \
    tar \
    --no-install-recommends && rm -rf /var/lib/apt/lists/*

# Download and extract Apache Flink
RUN curl -L https://downloads.apache.org/flink/flink-${FLINK_VERSION}/flink-${FLINK_VERSION}-bin-scala_${SCALA_VERSION}.tgz \
    | tar -xz -C /opt && mv /opt/flink-${FLINK_VERSION} $FLINK_HOME

# Set Flink binaries to PATH
ENV PATH="$FLINK_HOME/bin:$PATH"

# Set up working directory
WORKDIR $FLINK_HOME

# Expose Flink's default ports
EXPOSE 8081 6123 6124

# Copy entrypoint script (optional, create a custom entrypoint script if needed)
COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Define default command
CMD ["help"]