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

   > You may need to add `traefik.localhost`, `ollama.localhost`, and `webui.localhost` to your hosts file (`127.0.0.1`).

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