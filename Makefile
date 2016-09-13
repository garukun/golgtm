ifndef BUILD_SCOPE
BUILD_SCOPE=dev
endif

PROJECT=github.com/garukun/golgtm
PROJECT_IMAGE=garukun/golgtm:$(BUILD_SCOPE)

GO_IMAGE=vungle/golang:1.7

DOCKER_GOPATH=$(shell docker run --rm $(GO_IMAGE) /bin/bash -c 'echo $$GOPATH')
DOCKER_WORKDIR=$(DOCKER_GOPATH)/src/$(PROJECT)
BUILDDIR=server/httpserver
DOCKER_BUILD_SHELL=\
docker run --rm \
-v $$(pwd):$(DOCKER_WORKDIR) \
-v $$(pwd)/_out:/out \
-e CGO_ENABLED=0 \
-w $(DOCKER_WORKDIR)/$(BUILDDIR) \
$(SHELL_OPTS) \
$(GO_IMAGE)
DOCKER_TEST_SHELL=\
docker run --rm \
-v $$(pwd):$(DOCKER_WORKDIR) \
-v $$(pwd)/_out:/out \
-w $(DOCKER_WORKDIR) \
$(SHELL_OPTS) \
$(GO_IMAGE)

lint:
	@go fmt $(go list ./... | grep -v vendor)
	@go vet $(go list ./... | grep -v vendor)
	@$(DOCKER_BUILD_SHELL) golint | egrep -v "vendor"

deps: clean
ifeq ($(LATEST),true)
	rm $(BUILDDIR)/glide.lock
endif
	@echo "Vendoring external dependencies"
	@$(DOCKER_BUILD_SHELL) glide install
	@echo "All deps good!"

dev:
	$(eval SHELL_OPTS := -it)
	@$(DOCKER_TEST_SHELL) /bin/bash

ifndef TESTS
TESTS=./...
endif
test:
	@$(DOCKER_TEST_SHELL) go test $(TEST_OPTS) $$(go list $(TESTS) | grep -v vendor)

build:
	@$(DOCKER_BUILD_SHELL) go build \
	-a \
	-ldflags "-s -X main.revision=`git rev-parse HEAD`" \
	-o /out/golgtm
	@docker build -t $(PROJECT_IMAGE) .

publish:
	@docker push $(PROJECT_IMAGE)

clean:
	@rm -rf $(BUILDDIR)/_out $(BUILDDIR)/vendor
	@docker volume rm $$(docker volume ls -qf dangling=true) > /dev/null 2>/dev/null || true
	@echo "Cleaned!"
