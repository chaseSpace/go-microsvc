.PHONY: image pb new k8syaml bin up importsql help deploy

SVC ?=
ENV ?= 'beta'
TAG ?= '1.0.0'
BRANCH ?= ''
K8S ?= 0 # 是否k8s架构

DIR ?=

# 定义一个变量，列出所有目标
ALL_TARGETS := all image pb new k8syaml bin deploy update importsql

define required_svc
   $(if $1,,$(error SVC is required))
endef

define required_dir
   $(if $1,,$(error DIR is required))
endef

# 默认目标
help:
	$(info Usage:)
	$(info  make [subcommand])
	$(info)
	$(info subcommand:)
	$(foreach target,$(ALL_TARGETS),$(info     $(target))$(NEWLINE))

image:
	$(info Building image for SVC=$(SVC) ENV=$(ENV) TAG=$(TAG))
	@sh ./shellscript/build_image.sh $(SVC) $(ENV) $(TAG)

pb:
	@sh ./shellscript/build_pb.sh $(SVC)

new:
	$(info Initializing new microservice SVC=$(SVC))
	@sh ./shellscript/new_svc.sh $(SVC)

k8syaml:
	$(info Generating k8s yaml for SVC=$(SVC) ENV=$(ENV))
	@sh ./shellscript/gen_k8s_yaml.sh $(SVC) $(ENV)

install:
	$(info Installing golangci-lint)
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint
bin:
	$(info Building binary for SVC=$(SVC))
	@sh ./shellscript/build_bin.sh $(SVC) $(K8S)

deploy:
	$(info Deploying microservice SVC=$(SVC))
	@git pull && git submodule update && make pb SVC=$(SVC) && make bin SVC=$(SVC) && pm2 start pm2.config.js --only "$(SVC)" --env $(ENV)

up:
	$(info Updating microservice SVC=$(SVC))
	@git pull && git submodule update && make pb SVC=$(SVC) && make bin SVC=$(SVC) && pm2 reload pm2.config.js --only "$(SVC)" --env $(ENV)

rs:
	git reset origin/$(BRANCH) --hard

importsql:
	# yum install mysql -y
	$(call required_dir,$(DIR))
	$(info Importing SQL for DIR=$(DIR))
	@sh ./shellscript/import_sql.sh $(DIR)