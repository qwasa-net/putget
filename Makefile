SHELL := /bin/bash
TIMESTAMP := $(shell date +"%Y%m%d-%H%M")

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MYROOT := $(dir $(MAKEFILE_PATH))

DOCKER ?= docker
MENAME ?= putget
REGISTRY ?= qwasa.net/putget

DEPLOY_HOST := putget.qwasa.net
DEPLOY_PATH := /home/putget.qwasa.net/

#
.PHONY: help

help:
	-@grep -E "^[a-z_0-9]+:" "$(strip $(MAKEFILE_LIST))" | grep '##' | sed 's/:.*##/## —/ig' | column -t -s '##'

#
tea: build start  ## build and start

build: ## build
	CGO_ENABLED=0 go build -o "$(MYROOT)/$(MENAME)" "$(MYROOT)/src/main.go"

start: # start
	"$(MYROOT)/$(MENAME)"

#
docker_build_image: build ## build image
	$(DOCKER) build --force-rm --tag "$(REGISTRY)" --file "$(MYROOT)/deploy/Dockerfile" "$(MYROOT)"

docker_run_container: ## run container locally
	$(DOCKER) run --name "$(MENAME)" --publish 127.0.0.1:18801:18801 "$(REGISTRY)"

docker_push_image: ## push container
	$(DOCKER) push "$(REGISTRY):latest"

#
remote_load_image: ## load container (@deploy host)
	$(DOCKER) save "$(REGISTRY)" | ssh $(DEPLOY_HOST) $(DOCKER) load

remote_pull_image: ## pull container (@deploy host)
	$(DOCKER) pull "$(REGISTRY):latest"

remote_create_container: remote_stop_container ## create container from loaded/pulled image (@deploy host)
	ssh $(DEPLOY_HOST) '$(DOCKER) create \
	--name "$(MENAME)" \
	--publish 127.0.0.1:18801:18801 \
	--volume "$(DEPLOY_PATH)/files":/files \
	"$(REGISTRY)"'

remote_copy_files: ## copy files (to deploy host)
	ssh $(DEPLOY_HOST) 'mkdir -p "$(DEPLOY_PATH)" "$(DEPLOY_PATH)/files" "$(DEPLOY_PATH)/logs"'
	-scp -r ./deploy ./src ./Makefile ./keys/{_htpasswd,_env} "$(DEPLOY_HOST):$(DEPLOY_PATH)"

remote_stop_container:
	-@ssh $(DEPLOY_HOST) '$(DOCKER) container stop -t 1 $(MENAME); $(DOCKER) container rm $(MENAME)'

remote_install: ## setup nginx proxy and systemd service (@deploy host)
	ssh $(DEPLOY_HOST) sudo \
	"ln -sf $(DEPLOY_PATH)/deploy/nginx_putget.conf /etc/nginx/sites-enabled/; \
	systemctl reload nginx;\
	systemctl link --force $(DEPLOY_PATH)/deploy/systemd-podman-putget.service;\
	systemctl enable systemd-podman-putget.service;\
	systemctl status systemd-podman-putget.service;\
	systemctl stop systemd-podman-putget.service; sleep 2;\
	systemctl start systemd-podman-putget.service"

remote_restart:
	ssh $(DEPLOY_HOST) sudo \
	"systemctl reload nginx;\
	systemctl stop systemd-podman-putget.service; sleep 2;\
	systemctl start systemd-podman-putget.service"
