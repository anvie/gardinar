OUTPUT_DIR=output
BIN_TARGETS=$(addprefix $(OUTPUT_DIR)/, gardinar-darwin-amd64 gardinar-darwin-arm64 gardinar-linux-amd64 gardinar-linux-arm64)
TARGETS=$(BIN_TARGETS:=.zip)

all: $(TARGETS)

define build_binary
$(OUTPUT_DIR)/gardinar-$1: GOOS=$2 GOARCH=$3
$(OUTPUT_DIR)/gardinar-$1: main.go
	mkdir -p $(OUTPUT_DIR)
	GOOS=$$(GOOS) $$(GOARCH) go build -o $$@ main.go

$(OUTPUT_DIR)/gardinar-$1.zip: $(OUTPUT_DIR)/gardinar-$1
	cd $(OUTPUT_DIR) && zip $$(notdir $$@) gardinar-$1
endef

$(eval $(call build_binary,darwin-amd64,darwin,amd64))
$(eval $(call build_binary,darwin-arm64,darwin,arm64))
$(eval $(call build_binary,linux-amd64,linux,amd64))
$(eval $(call build_binary,linux-arm64,linux,arm64))

clean:
	rm -f output/*

.PHONY: clean all