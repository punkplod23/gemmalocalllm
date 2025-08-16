# gemmalocalllm

Simple local LLM stack using Docker and Traefik.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) (latest version recommended)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Rancher Desktop](https://rancherdesktop.io/) (optional, for local Kubernetes clusters)
- [Traefik](https://doc.traefik.io/traefik/) (set up via Docker Compose)
- [Chocolatey](https://chocolatey.org/install) (optional, for Windows package management)
- At least 8GB RAM recommended for LLMs
- Download the desired LLM model (e.g., Gemma) and place it in a directory accessible to Docker

### NVIDIA Container Runtime (for GPU support)

#### Linux

1. Install NVIDIA drivers for your GPU.
2. Install the NVIDIA Container Toolkit:
   ```sh
   sudo apt-get update
   sudo apt-get install -y nvidia-driver-535
   sudo systemctl restart docker
   distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
   curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | sudo apt-key add -
   curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
   sudo apt-get update
   sudo apt-get install -y nvidia-container-toolkit
   sudo systemctl restart docker
   ```
3. Test with:
   ```sh
   docker run --rm --gpus all nvidia/cuda:12.0.0-base-ubuntu20.04 nvidia-smi
   ```
4. **(Optional but recommended)**: Set the NVIDIA runtime as default in Docker.
   Edit or create `/etc/docker/daemon.json` and add:
   ```json
   {
       "runtimes": {
           "nvidia": {
               "path": "/usr/bin/nvidia-container-runtime",
               "runtimeArgs": []
           }
       },
       "default-runtime": "nvidia"
   }
   ```
   Then restart Docker:
   ```sh
   sudo systemctl restart docker
   ```

#### Windows (WSL2)

- Install [NVIDIA drivers](https://www.nvidia.com/Download/index.aspx) for your GPU.
- Install [WSL2](https://docs.microsoft.com/en-us/windows/wsl/install) and ensure you have Ubuntu or another supported distro.
- Install Docker Desktop and enable WSL2 integration.
- In Docker Desktop, go to **Settings > Resources > WSL Integration** and enable integration for your distro.
- Enable GPU support in Docker Desktop: **Settings > Resources > GPU** (check "Enable GPU support").
- (Optional) To ensure the NVIDIA runtime is set, you can add the following to your WSL2's `/etc/docker/daemon.json` (inside your Linux distro):
   ```json
   {
       "runtimes": {
           "nvidia": {
               "path": "/usr/bin/nvidia-container-runtime",
               "runtimeArgs": []
           }
       },
       "default-runtime": "nvidia"
   }
   ```
   Then restart Docker inside WSL2:
   ```sh
   sudo service docker restart
   ```
- Test with:
   ```sh
   docker run --rm --gpus all nvidia/cuda:12.0.0-base-ubuntu20.04 nvidia-smi
   ```

> For more details, see the [NVIDIA Container Toolkit documentation](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html).

## Quick Start

1. **Clone this repository:**
   ```sh
   git clone https://github.com/punkplod23/gemmalocalllm.git
   cd gemmalocalllm
   ```

2. **Create the Docker network:**
   ```sh
   docker network create llm-web-app-network
   ```

3. **Start the stack:**
   ```sh
   docker compose up -d --build
   ```

4. **Run the Gemma model in Ollama:**
   ```sh
   docker exec -it ollama ollama run gemma:2b
   ```

5. **Access services:**
   - Traefik dashboard: [https://traefik.localhost](https://traefik.localhost)
   - Ollama API: [https://ollama.localhost](https://ollama.localhost)
   - Open WebUI: [https://webui.localhost](https://webui.localhost)
   - Adminer (Postgres UI): [https://adminer.localhost](https://adminer.localhost)

   > You may need to add `traefik.localhost`, `ollama.localhost`, `webui.localhost`, and `adminer.localhost` to your hosts file (`127.0.0.1`).

6. **Test CUDA GPU support (optional):**
   ```sh
   docker exec -it cuda nvidia-smi
   ```
   This will show your GPU info inside the CUDA 11.4.0 container.

## Loading CSV Files into Postgres

1. Place your CSV files in the `csv` directory at the root of this project (`./csv`).
2. Connect to the Postgres container:
   ```sh
   docker exec -it postgres psql -U ollama -d ollama
   ```
3. In the `psql` prompt, create a table matching your CSV structure, for example:
   ```sql
   CREATE TABLE my_vectors (
     id SERIAL PRIMARY KEY,
     text TEXT,
     embedding VECTOR(1536) -- adjust dimension as needed
   );
   ```
4. Load your CSV file (assuming columns match):
   ```sql
   \copy my_vectors(text) FROM '/csv/yourfile.csv' DELIMITER ',' CSV HEADER;
   ```
5. You can then use the `vector` extension for similarity search, etc.

> The `/csv` directory is mounted read-only into the Postgres container at `/csv`.

## Notes

- Ensure your user has permission to run Docker commands.
- For HTTPS with real certificates, configure your email in `docker-compose.yml` and uncomment the Let's Encrypt lines.
- For production, secure the Traefik dashboard and restrict ports as needed.
- GPU is recommended for best performance, but CPU can be used by removing the GPU device section in `docker-compose.yml`.

## Troubleshooting

- If you encounter network errors, ensure the `llm-web-app-network` exists and is external.
- Check container logs with `docker logs <container_name>` for more details.

## example
curl -k -X POST \ https://ollama.localhost/api/generate \ -H "Content-Type: application/json" \ -d '{   "model": "gemma:2b",    "prompt": "Tell me a short, funny story about a talking cat.",  "stream": false}'

## Decision
not work the time