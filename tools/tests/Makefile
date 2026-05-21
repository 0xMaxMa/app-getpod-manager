# getpod-manager API test scripts
# Copy .env.example to .env and fill in values before running.
#
# BASE_URL defaults to http://localhost:5990 (direct)
# Set BASE_URL=http://localhost:10850/app/getpod-manager/api to test via gateway

SHELL := /bin/bash
DIR   := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

.PHONY: health metrics resize ssh-keys-list ssh-keys-add ssh-keys-delete help

help:
	@echo "Available targets:"
	@echo "  make health              GET /health (no auth)"
	@echo "  make metrics             GET /metrics"
	@echo "  make resize              POST /resize  (DISK_GIB= CPU_CORES= MEMORY_MIB=)"
	@echo "  make ssh-keys-list       GET /ssh-keys"
	@echo "  make ssh-keys-add        POST /ssh-keys  (SSH_KEY=)"
	@echo "  make ssh-keys-delete     DELETE /ssh-keys/:fingerprint  (FINGERPRINT=)"
	@echo ""
	@echo "Example:"
	@echo "  make metrics"
	@echo "  DISK_GIB=50 make resize"
	@echo "  SSH_KEY='ssh-ed25519 AAAA...' make ssh-keys-add"
	@echo "  FINGERPRINT='SHA256:...' make ssh-keys-delete"

health:
	@bash $(DIR)health.sh

metrics:
	@bash $(DIR)metrics.sh

resize:
	@bash $(DIR)resize.sh

ssh-keys-list:
	@bash $(DIR)ssh-keys-list.sh

ssh-keys-add:
	@bash $(DIR)ssh-keys-add.sh

ssh-keys-delete:
	@bash $(DIR)ssh-keys-delete.sh
