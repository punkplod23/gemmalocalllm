# Use the official Unsloth image with CUDA 12.1 and Python 3.10
FROM unsloth/unsloth:latest-cu121-py3.10

# Set the working directory
WORKDIR /app

# Copy the training script into the container
COPY train.py .

# Install necessary Python packages for training
# Using a requirements.txt file is also a good practice here.
RUN pip install "unsloth[cu121] @ git+https://github.com/unslothai/unsloth.git"
RUN pip install "tyro" "datasets" "trl" "transformers" "accelerate" "bitsandbytes"

# The command to run when the container starts will be provided by the docker-compose.yml file.
CMD ["python", "train.py"]