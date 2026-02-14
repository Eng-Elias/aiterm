FROM docker.litellm.ai/berriai/litellm:main-stable

WORKDIR /app

# Copy proxy configuration
COPY config.yaml .

EXPOSE 4000

# UI is enabled by default at /ui — no --ui flag needed
CMD ["--port", "4000", "--config", "config.yaml"]
