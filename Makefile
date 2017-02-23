ASSETS_FOLDER := assets
ASSETS := $(shell ls $(ASSETS_FOLDER))
BINDATA_CMD := go-bindata
BINDATA_FILE := bindata.go

bindata: $(ASSETS)

$(ASSETS):
	$(BINDATA_CMD) \
		-pkg $@ -ignore $(BINDATA_FILE) \
		-prefix $(ASSETS_FOLDER)/$@ \
		-o $(ASSETS_FOLDER)/$@/$(BINDATA_FILE) \
		assets/$@/...