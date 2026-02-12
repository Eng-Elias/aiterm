FROM ghcr.io/berriai/litellm:main-latest

# Copy proxy configuration
COPY config.yaml /app/config.yaml

# Health check — LiteLLM exposes /health
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["curl", "-f", "http://localhost:4000/health"] || exit 1

EXPOSE 4000

# Start LiteLLM proxy with config
CMD ["litellm", "--config", "/app/config.yaml", "--port", "4000"]
