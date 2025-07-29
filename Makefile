LAMBDAS := server_status
BIN_DIR := bin
DIST_DIR := dist

.PHONY: all $(LAMBDAS) clean

all: $(LAMBDAS)

$(LAMBDAS):
	@echo "==> Building $@"
	GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$@/$@ ./$(shell echo $@)/.
	@echo "==> Zipping $@"
	mkdir -p $(DIST_DIR)/$@
	cd $(BIN_DIR)/$@ && zip -q ../../$(DIST_DIR)/$@/$@.zip $@

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)
