ASSETS_PATH := etc
ASSETS := $(shell ls $(ASSETS_PATH))
ASSETS_PACKAGE := assets
BINDATA_FILE := bindata.go
BINDATA_CMD := go-bindata

all: bindata

bindata: $(ASSETS)

$(ASSETS):
	$(BINDATA_CMD) \
		-pkg $@ \
		-prefix $(ASSETS_PATH)/$@ \
		-o $(ASSETS_PACKAGE)/$@/$(BINDATA_FILE) \
		$(ASSETS_PATH)/$@/...