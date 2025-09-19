KUBECTL ?= kubectl
K8S_DIR := $(CURDIR)

all: apply

apply:
	$(KUBECTL) apply -f postgres.yaml
	$(KUBECTL) apply -f ollama.yaml
	$(KUBECTL) apply -f chromadb.yaml
	$(KUBECTL) apply -f open-webui.yaml
	$(KUBECTL) apply -f adminer.yaml
	$(KUBECTL) apply -f ollama-ingress.yaml
	$(KUBECTL) apply -f open-webui-ingress.yaml
	$(KUBECTL) apply -f chromadb-ingress.yaml
	$(KUBECTL) apply -f adminer-ingress.yaml

clean:
	$(KUBECTL) delete -f postgres.yaml --ignore-not-found
	$(KUBECTL) delete -f ollama.yaml --ignore-not-found
	$(KUBECTL) delete -f ollama-ingress.yaml --ignore-not-found
	$(KUBECTL) delete -f chromadb.yaml --ignore-not-found
	$(KUBECTL) delete -f open-webui.yaml --ignore-not-found
	$(KUBECTL) delete -f open-webui-ingress.yaml --ignore-not-found
	$(KUBECTL) delete -f chromadb-ingress.yaml --ignore-not-found
	$(KUBECTL) delete -f adminer.yaml --ignore-not-found
	$(KUBECTL) delete -f adminer-ingress.yaml --ignore-not-found


status:
	$(KUBECTL) get all

loadlama:
	$(KUBECTL) apply -f ollama-model-loader.yaml

pull-gemma3-270m:
	$(KUBECTL) exec -it deploy/ollama -- ollama pull gemma3:270

pull-gemma-code:
	$(KUBECTL) exec -it deploy/ollama -- ollama pull codegemma:2b

pull-vision:
	$(KUBECTL) exec -it deploy/ollama -- ollama pull gemma3:4b


