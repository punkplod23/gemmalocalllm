install-network:
	docker network create llm-web-app-network

install:
	docker compose up -d --build

run:
	docker exec -it ollama ollama run gemma:2b	

