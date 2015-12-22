DEPSDIR = /deps
SRCDIR = /src/github.com/garukun/golgtm
GOLANG_IMAGE = vungle/golang
SRC_IMAGE = garukun/golgtm

default: build

deps:
	@docker pull $(GOLANG_IMAGE)
	$(eval GOPATH := $(shell docker run --rm $(GOLANG_IMAGE) /bin/bash -c 'echo $$GOPATH'))
	$(eval OUTDIR := $(shell docker run --rm $(GOLANG_IMAGE) /bin/bash -c 'echo $$OUTDIR'))
	@rm -rf ./vendor
	docker run --rm \
		-v `pwd`:$(DEPSDIR) \
		-w $(DEPSDIR) \
		$(GOLANG_IMAGE) \
		glide up

test:
	$(error Testing for development is not implemented!)

test-ci: deps
	docker run --rm \
		-v `pwd`:$(GOPATH)$(SRCDIR) \
		-v `pwd`/_out:$(OUTDIR) \
		-w $(GOPATH)$(SRCDIR) \
		$(GOLANG_IMAGE) \
		/bin/bash -c "coverage.sh | report.sh"

build: deps
	@rm -rf _out/golgtm
	docker run --rm \
		-v `pwd`:$(GOPATH)$(SRCDIR) \
		-v `pwd`/_out:$(OUTDIR) \
		-w $(GOPATH)$(SRCDIR) \
		-e "CGO_ENABLED=0" \
		$(GOLANG_IMAGE) \
		go build -a -ldflags '-s' -o $(OUTDIR)/golgtm

build-prod:
	$(eval PROD_IMAGE := $(shell docker build -q -t $(SRC_IMAGE) . | awk '/Successfully built/{print $$NF}'))
	$(if $(PROD_IMAGE), @echo $(PROD_IMAGE), $(error Cannot build production image))

tag-docker: build-prod
	docker tag -f $(PROD_IMAGE) $(SRC_IMAGE)
	docker tag -f $(PROD_IMAGE) $(SRC_IMAGE):$(TAG)

run: build build-prod
	docker run $(PROD_IMAGE)

bench:
	$(error Benchmarking is not implemented!)

nuke:
	-rm -rf _out vendor
	-docker rm -f `docker ps -aq`
	-docker rmi -f `docker images -aq`
