-include .sdk/Makefile

$(if $(filter true,$(sdkloaded)),,$(error You must install bblfsh-sdk))

test-native-internal:
	cd native; \
	echo "not implemented"

build-native-internal:
	cd native; \
	echo "not implemented"
	echo -e "#!/bin/bash\necho 'not implemented'" > $(BUILD_PATH)/native
	chmod +x $(BUILD_PATH)/native
