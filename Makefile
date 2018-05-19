# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_DIR=bin/$(PLATFORM)/$(ARCH)
SRC_DIR=src
BINARY_NAME=docker-init

all: build
build:
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) -v ./$(SRC_DIR)
clean:
	$(GOCLEAN)
	rm -f $(BINARY_DIR)/$(BINARY_NAME)
run:
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) -v ./$(SRC_DIR)
	./$(BINARY_DIR)/$(BINARY_NAME)
