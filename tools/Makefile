.PHONY: help install uninstall stat job install-watch start stop logs

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

install: ## Install app via gateway (returns jobId)
	bash install.sh

uninstall: ## Uninstall app via gateway
	bash uninstall.sh

stat: ## Show app status
	bash stat.sh

job: ## Poll job status  usage: make job id=<jobId>
	bash job.sh $(id) | jq

install-watch: ## Install app and poll until done
	@JOB=$$(bash install.sh | jq -r '.jobId') && \
	echo "jobId: $$JOB" && \
	while true; do \
		RESULT=$$(bash job.sh $$JOB); \
		printf '%s' "$$RESULT" | jq .; \
		printf '%s' "$$RESULT" | jq -e '.status == "completed" or .status == "failed"' > /dev/null 2>&1 && break; \
		sleep 2; \
	done

start: ## Start app with docker compose
	docker compose -f ../docker-compose.yml up -d

stop: ## Stop app with docker compose
	docker compose -f ../docker-compose.yml down

logs: ## Follow app logs
	docker compose -f ../docker-compose.yml logs -f
