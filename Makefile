# Variables
IMAGE_NAME   ?= http-server
IMAGE_TAG    ?= v1.0.0
CLUSTER_NAME ?= local-dev
MANIFEST     ?= ./k8s/

# ANSI Color Codes
CLR_RESET   := \033[0m
CLR_BOLD    := \033[1m
CLR_CYAN    := \033[36m
CLR_GREEN   := \033[32m
CLR_YELLOW  := \033[33m
CLR_BLUE    := \033[34m
CLR_MAGENTA := \033[35m

.PHONY: all build load apply rollout deploy logs status test port-forward clean

# Default target: builds, loads into kind, applies manifests, and waits for rollout
deploy: build load apply rollout status

# 1. Build the Docker image
build:
	@printf "$(CLR_CYAN)$(CLR_BOLD)==> Building Docker image %s:%s...$(CLR_RESET)\n" "$(IMAGE_NAME)" "$(IMAGE_TAG)"
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# 2. Load the Docker image directly into the Kind cluster
load:
	@printf "$(CLR_BLUE)$(CLR_BOLD)==> Loading image into Kind cluster '%s'...$(CLR_RESET)\n" "$(CLUSTER_NAME)"
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name $(CLUSTER_NAME)

# 3. Apply Kubernetes YAML manifests
apply:
	@printf "$(CLR_YELLOW)$(CLR_BOLD)==> Applying Kubernetes manifests from %s...$(CLR_RESET)\n" "$(MANIFEST)"
	kubectl apply -f $(MANIFEST)

# 4. Restart deployment and wait for zero-downtime rollout
rollout:
	@printf "$(CLR_MAGENTA)$(CLR_BOLD)==> Restarting deployment to pick up new image...$(CLR_RESET)\n"
	kubectl rollout restart deployment/http-server
	kubectl rollout status deployment/http-server --timeout=60s

# Show current pods, services, and ingress resources
status:
	@printf "$(CLR_GREEN)$(CLR_BOLD)==> Cluster status:$(CLR_RESET)\n"
	@kubectl get pods,svc,ingress -l app=http-server

# Stream logs from the server pods
logs:
	kubectl logs -l app=http-server -f --max-log-requests=10

# Test the Ingress endpoint
test:
	@printf "$(CLR_GREEN)$(CLR_BOLD)==> Sending test request via Ingress...$(CLR_RESET)\n"
	curl -i -H "Host: myserver.local" http://localhost/

# Port-forward service directly to localhost:8080 (bypass Ingress for debugging)
port-forward:
	kubectl port-forward svc/http-server-svc 8080:80

# Clean up deployed resources
clean:
	@printf "$(CLR_YELLOW)$(CLR_BOLD)==> Deleting manifests in %s...$(CLR_RESET)\n" "$(MANIFEST)"
	kubectl delete -f $(MANIFEST) --ignore-not-found