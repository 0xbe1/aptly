.PHONY: build release-patch release-minor release-major clean

BINARY := aptly
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT_SHA := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

VERSION_PARTS := $(subst ., ,$(subst v,,$(CURRENT_VERSION)))
MAJOR := $(word 1,$(VERSION_PARTS))
MINOR := $(word 2,$(VERSION_PARTS))
PATCH := $(word 3,$(VERSION_PARTS))

build:
	APTLY_VERSION=$(CURRENT_VERSION) \
	APTLY_GIT_SHA=$(COMMIT_SHA) \
	APTLY_BUILD_DATE=$(BUILD_DATE) \
	cargo build -p aptly-cli --release --bin $(BINARY)
	cp target/release/$(BINARY) ./$(BINARY)

release-patch:
	@NEW_VERSION=v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))); \
	echo "Releasing $$NEW_VERSION (was $(CURRENT_VERSION))"; \
	git tag $$NEW_VERSION && git push origin $$NEW_VERSION

release-minor:
	@NEW_VERSION=v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0; \
	echo "Releasing $$NEW_VERSION (was $(CURRENT_VERSION))"; \
	git tag $$NEW_VERSION && git push origin $$NEW_VERSION

release-major:
	@NEW_VERSION=v$(shell echo $$(($(MAJOR)+1))).0.0; \
	echo "Releasing $$NEW_VERSION (was $(CURRENT_VERSION))"; \
	git tag $$NEW_VERSION && git push origin $$NEW_VERSION

clean:
	rm -f $(BINARY)
	cargo clean
