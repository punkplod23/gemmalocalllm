install:
	docker network create llm-web-app-network
	docker compose -f 'docker-compose.yml' up -d --build
	
run:
	docker exec -it ollama ollama run gemma:2b	