# Directories
FUNCTIONS := $(wildcard functions/*)
BUILD_DIR := build

# Extract function names from folder names
FUNCTION_NAMES := $(notdir $(FUNCTIONS))

# Default target: build all
.PHONY: all
all: $(FUNCTION_NAMES)

# Rule for each function
$(FUNCTION_NAMES):
	@echo "Building $@..."
	@mkdir -p $(BUILD_DIR)/$@
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BUILD_DIR)/$@/bootstrap ./functions/$@
	@cd $(BUILD_DIR)/$@ && zip -j ../$@.zip bootstrap

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)/*
