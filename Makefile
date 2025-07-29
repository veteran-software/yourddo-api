LAMBDAS := server_status
BIN_DIR := bin
DIST_DIR := dist

.PHONY: all $(LAMBDAS) clean

all: $(LAMBDAS)

$(LAMBDAS):
	@echo "==> Building $@"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$@/bootstrap ./$(shell echo $@)/.
	@echo "==> Zipping $@"
	mkdir -p $(DIST_DIR)/$@
	cd $(BIN_DIR)/$@ && zip -q ../../$(DIST_DIR)/$@/$@.zip bootstrap

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)
