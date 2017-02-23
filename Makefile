ASSETS_PATH := etc
ASSETS_PACKAGE := assets
ASSETS := $(shell ls $(ASSETS_PATH))
BINDATA_CMD := go-bindata

bindata: $(ASSETS)

$(ASSETS):
	$(BINDATA_CMD) \
		-pkg $@ \
		-prefix $(ASSETS_PATH)/$@ \
		-o $(ASSETS_PACKAGE)/$@/$(BINDATA_FILE) \
		$(ASSETS_PATH)/$@/...