SHELL := /bin/bash
TIMESTAMP := $(shell date +"%Y%m%d-%H%M")

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
MYROOT := $(dir $(MAKEFILE_PATH))

GOPATH := $(MYROOT)/_go
GO_BUILD_CGO_ENABLED := 1
GO_BUILD_OPTS := -a -ldflags "-linkmode external -extldflags '-static'" -tags netgo,sqlite_omit_load_extension
GO_BUILD_OPTS_DEV :=

DOCKER ?= docker
MENAME ?= putget
REGISTRY ?= qwasa.net/putget

DEPLOY_USER := putget.qwasa.net
DEPLOY_USER_KEY := ./_keys/id_rsa_putget
DEPLOY_USER_KEY_PUB := ./_keys/id_rsa_putget.pub
DEPLOY_PATH := /home/$(DEPLOY_USER)
DEPLOY_HOST := putget.qwasa.net
DEPLOY_TARGET := $(DEPLOY_USER)@$(DEPLOY_HOST)
SSH := ssh -i $(DEPLOY_USER_KEY)
SCP := scp -i $(DEPLOY_USER_KEY)

#
.PHONY: help

help:
	-@grep -E "^[a-z_0-9]+:" "$(strip $(MAKEFILE_LIST))" | grep '##' | sed 's/:.*##/## —/ig' | column -t -s '##'
	-@echo
	-@echo ' ---'
	-@echo ' make tea'
	-@echo ' ---'
	-@echo ' make remote_sudo_install DEPLOY_TARGET=root@putget.qwasa.net # root access required'
	-@echo ' make docker_build_image remote_load_image remote_copy_files remote_create_container'
	-@echo ' make remote_install_service remote_restart_service'
	-@echo

#
tea: build start ## build and start
instant_coffee: build_dev start # build fast and start

build: goget ## build
	cd "$(MYROOT)/src/" && \
	GOPATH="$(GOPATH)" \
	CGO_ENABLED=$(GO_BUILD_CGO_ENABLED) \
	go build -C "$(MYROOT)/src/" $(GO_BUILD_OPTS) -o "$(MYROOT)/$(MENAME)" .

build_dev: GO_BUILD_OPTS = $(GO_BUILD_OPTS_DEV) # fast build
build_dev: build

goget: # get dependencies
	mkdir -vp "$(GOPATH)"
	cd "$(MYROOT)/src/" && GOPATH="$(GOPATH)" go mod vendor -v

start: ## start
	"$(MYROOT)/$(MENAME)"

tests: ## tests (NotImplementedYet)
	@echo "I will add some tests maybe later (most likely never)"

#
docker_build_image: build ## build image
	$(DOCKER) build --force-rm --tag "$(REGISTRY)" --file "$(MYROOT)/deploy/Dockerfile" "$(MYROOT)"

docker_run_container: ## run container locally
	$(DOCKER) run --rm --replace --name "$(MENAME)" --publish 127.0.0.1:18801:18801 "$(REGISTRY)"

docker_push_image: ## push container
	$(DOCKER) push "$(REGISTRY):latest"

#
remote_load_image: ## load container (@deploy host)
	$(DOCKER) save "$(REGISTRY)" | $(SSH) $(DEPLOY_TARGET) $(DOCKER) load

remote_pull_image: ## pull container (@deploy host)
	$(DOCKER) pull "$(REGISTRY):latest"

remote_create_container: remote_stop_container ## create container from loaded/pulled image (@deploy host)
	$(SSH) $(DEPLOY_TARGET) '$(DOCKER) create \
	--name "$(MENAME)" \
	--network=slirp4netns \
	--publish 127.0.0.1:18801:18801 \
	--volume "$(DEPLOY_PATH)/files":/files \
	"$(REGISTRY)"'

remote_copy_files: ## copy files (to deploy host)
	$(SSH) $(DEPLOY_TARGET) 'whoami; mkdir -vp "$(DEPLOY_PATH)" "$(DEPLOY_PATH)/files" "$(DEPLOY_PATH)/logs"'
	-$(SCP) -r ./deploy ./misc/www "$(DEPLOY_TARGET):$(DEPLOY_PATH)"
	-[ -f ./_keys/htpasswd ] && $(SCP) -r ./_keys/htpasswd "$(DEPLOY_TARGET):$(DEPLOY_PATH)"
	-[ -d ./_keys/deploy ] && $(SCP) -r ./_keys/deploy "$(DEPLOY_TARGET):$(DEPLOY_PATH)"

remote_stop_container:
	-@$(SSH) $(DEPLOY_TARGET) '$(DOCKER) container stop -t 1 $(MENAME); $(DOCKER) container rm $(MENAME)'

create_keys: HT_USER := $(shell openssl rand -hex 16)
create_keys: HT_PASSWD := $(shell openssl rand -hex 16)
create_keys:
	mkdir -pv ./_keys
	[ -f ./_keys/id_rsa_putget ] || \
	ssh-keygen -t rsa -b 4096 -f ./_keys/id_rsa_putget -C "putget"
	[ -f ./_keys/htpasswd ] || \
	(htpasswd -Bcb ./_keys/htpasswd "$(HT_USER)" "$(HT_PASSWD)" && \
	echo "$(HT_USER):$(HT_PASSWD)" | tee ./_keys/htpasswd.txt)

remote_sudo_install: create_keys ## create service user (@deploy host)
	$(SSH) $(DEPLOY_TARGET) sudo "\
	useradd --user-group --groups www-data --shell /bin/bash --create-home --home-dir $(DEPLOY_PATH) $(DEPLOY_USER); \
	loginctl enable-linger $(DEPLOY_USER); \
	chmod 711 $(DEPLOY_PATH); \
	mkdir -pv $(DEPLOY_PATH)/.ssh; \
	"
	-$(SCP) -r "$(DEPLOY_USER_KEY_PUB)" "$(DEPLOY_TARGET):$(DEPLOY_PATH)/.ssh/authorized_keys"
	$(SSH) $(DEPLOY_TARGET) sudo "\
	chown -R $(DEPLOY_USER):$(DEPLOY_USER) $(DEPLOY_PATH)/.ssh; \
	chmod 700 $(DEPLOY_PATH)/.ssh; \
	chmod 600 $(DEPLOY_PATH)/.ssh/authorized_keys; \
	ln -sf $(DEPLOY_PATH)/deploy/nginx_putget.conf /etc/nginx/sites-enabled/; \
	"

remote_install_service: ## setup nginx proxy and systemd service (@deploy host)
	$(SSH) $(DEPLOY_TARGET) "\
	systemctl --user link --force $(DEPLOY_PATH)/deploy/systemd-podman-putget.service; \
	systemctl --user link --force $(DEPLOY_PATH)/deploy/systemd-putget-maintenance.service; \
	systemctl --user link --force $(DEPLOY_PATH)/deploy/systemd-putget-maintenance.timer; \
	systemctl --user daemon-reload; \
	systemctl --user enable systemd-podman-putget.service; \
	systemctl --user enable systemd-putget-maintenance.timer; \
	systemctl --user enable systemd-putget-maintenance.service; \
	"

remote_stop_service:
	$(SSH) $(DEPLOY_TARGET) "\
	systemctl --user stop systemd-podman-putget.service; \
	"

remote_restart_service: remote_stop_service ## (re)start service (@deploy host)
	sleep 2
	$(SSH) $(DEPLOY_TARGET) "\
	systemctl --user daemon-reload; \
	systemctl --user start systemd-podman-putget.service; \
	systemctl --user status systemd-podman-putget.service; \
	sg www-data -c 'sudo systemctl restart nginx'; \
	"
