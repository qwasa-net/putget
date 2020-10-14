SHELL := /bin/bash
TIMESTAMP := $(shell date +"%Y%m%d-%H%M")

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MYROOT := $(dir $(MAKEFILE_PATH))

GOPATH := $(MYROOT)/vendor
GO_BUILD_CGO_ENABLED := 1
GO_BUILD_OPTS := -a -ldflags "-linkmode external -extldflags '-static'" -tags netgo,sqlite_omit_load_extension
GO_BUILD_OPTS_DEV :=

DOCKER ?= docker
MENAME ?= putget
REGISTRY ?= qwasa.net/putget

DEPLOY_HOST := putget.qwasa.net
DEPLOY_PATH := /home/putget.qwasa.net/

#
.PHONY: help

help:
	-@grep -E "^[a-z_0-9]+:" "$(strip $(MAKEFILE_LIST))" | grep '##' | sed 's/:.*##/## —/ig' | column -t -s '##'
	-@echo
	-@echo ' ---'
	-@echo ' make tea'
	-@echo ' ---'
	-@echo ' make build docker_build_image remote_load_image remote_create_container remote_copy_files DOCKER=podman'
	-@echo ' make remote_install DEPLOY_HOST=root@putget.qwasa.net'
	-@echo

#
tea: build start ## build and start
instant_coffee: build_dev start # build fast and start

build: goget ## build
	GOPATH="$(GOPATH)" \
	CGO_ENABLED=$(GO_BUILD_CGO_ENABLED) \
	go build $(GO_BUILD_OPTS) -o "$(MYROOT)/$(MENAME)" "$(MYROOT)/src/main.go"

build_dev: GO_BUILD_OPTS = $(GO_BUILD_OPTS_DEV) # fast build
build_dev: build

goget: # get dependencies
	mkdir -p "$(GOPATH)"
	cat "$(GOPATH)/packages.txt" | while read pkg; \
	do echo "$${pkg}"; GOPATH="$(GOPATH)" go get "$${pkg}"; done

start: ## start
	"$(MYROOT)/$(MENAME)"

#
docker_build_image: build ## build image
	$(DOCKER) build --force-rm --tag "$(REGISTRY)" --file "$(MYROOT)/deploy/Dockerfile" "$(MYROOT)"

docker_run_container: ## run container locally
	$(DOCKER) run --rm --replace --name "$(MENAME)" --publish 127.0.0.1:18801:18801 "$(REGISTRY)"

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
	-scp -r ./deploy ./src ./Makefile ./misc "$(DEPLOY_HOST):$(DEPLOY_PATH)"

remote_stop_container:
	-@ssh $(DEPLOY_HOST) '$(DOCKER) container stop -t 1 $(MENAME); $(DOCKER) container rm $(MENAME)'

remote_install: ## setup nginx proxy and systemd service (@deploy host)
	ssh $(DEPLOY_HOST) sudo \
	"ln -sf $(DEPLOY_PATH)/deploy/nginx_putget.conf /etc/nginx/sites-enabled/; \
	systemctl reload nginx;\
	systemctl link --force $(DEPLOY_PATH)/deploy/systemd-podman-putget.service;\
	systemctl link --force $(DEPLOY_PATH)/deploy/systemd-putget-maintenance.service;\
	systemctl link --force $(DEPLOY_PATH)/deploy/systemd-putget-maintenance.timer;\
	systemctl daemon-reload; \
	systemctl enable systemd-putget-maintenance.timer;\
	systemctl start systemd-putget-maintenance.timer;\
	systemctl enable systemd-podman-putget.service;\
	systemctl status systemd-podman-putget.service;\
	systemctl stop systemd-podman-putget.service; sleep 2;\
	systemctl start systemd-podman-putget.service"

remote_restart:
	ssh $(DEPLOY_HOST) sudo \
	"systemctl reload nginx;\
	systemctl daemon-reload; \
	systemctl stop systemd-podman-putget.service; sleep 2;\
	systemctl start systemd-podman-putget.service"
