KUBECTL ?= kubectl
K8S_DIR := $(CURDIR)

all: apply

apply:
	$(KUBECTL) apply -f https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/v0.14.5/nvidia-device-plugin.yml
	$(KUBECTL) apply -f letsencrypt-pvc.yaml
	$(KUBECTL) apply -f postgres.yaml
	$(KUBECTL) apply -f ollama.yaml
	$(KUBECTL) apply -f chromadb.yaml
	$(KUBECTL) apply -f open-webui.yaml
	$(KUBECTL) apply -f open-webui-ingress.yaml
	$(KUBECTL) apply -f pipelines.yaml
	$(KUBECTL) apply -f adminer.yaml
	$(KUBECTL) apply -f ollama-ingress.yaml

clean:
	$(KUBECTL) delete -f letsencrypt-pvc.yaml --ignore-not-found
	$(KUBECTL) delete -f postgres.yaml --ignore-not-found
	$(KUBECTL) delete -f ollama.yaml --ignore-not-found
	$(KUBECTL) delete -f chromadb.yaml --ignore-not-found
	$(KUBECTL) delete -f open-webui.yaml --ignore-not-found
	$(KUBECTL) delete -f open-webui-ingress.yaml --ignore-not-found
	$(KUBECTL) delete -f pipelines.yaml --ignore-not-found
	$(KUBECTL) delete -f adminer.yaml --ignore-not-found

status:
	$(KUBECTL) get all

loadlama:
	$(KUBECTL) apply -f ollama-model-loader.yaml