FROM ghcr.io/berriai/litellm:main-latest

# Copy configuration
COPY config.yaml /app/config.yaml

# Expose port
EXPOSE 7860

# Start with UI enabled
CMD ["litellm", "--config", "/app/config.yaml", "--port", "7860", "--ui"]
