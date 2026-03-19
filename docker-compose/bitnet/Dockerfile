# Stage 1: Build the C++ Backend and Download Model
FROM python:3.10-slim AS builder

RUN apt-get update && apt-get install -y \
    cmake \
    build-essential \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Clone the official BitNet inference framework
RUN git clone https://github.com/microsoft/BitNet.git .

# Install Python requirements
RUN pip install --no-cache-dir -r requirements.txt huggingface_hub

# Download the specific 1.58-bit GGUF model
RUN huggingface-cli download microsoft/BitNet-b1.58-2B-4T-gguf \
    --local-dir models/BitNet-b1.58-2B-4T

# Build the C++ backend
RUN python setup_env.py -md models/BitNet-b1.58-2B-4T -q i2_s

# Stage 2: Local Runtime
FROM python:3.10-slim

WORKDIR /app

RUN apt-get update && apt-get install -y libgomp1 && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app /app
RUN pip install --no-cache-dir -r requirements.txt

EXPOSE 8080

# Run the inference server (OpenAI API compatible)
ENTRYPOINT ["python", "run_inference_server.py"]
CMD ["-m", "models/BitNet-b1.58-2B-4T/ggml-model-i2_s.gguf", "--host", "0.0.0.0", "--port", "8080", "-t", "4", "-c", "2048"]
