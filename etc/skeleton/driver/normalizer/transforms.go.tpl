package normalizer

import "github.com/bblfsh/sdk/v3/driver"

var Transforms = driver.Transforms{
	Namespace:      "{{.Manifest.Language}}",
	Preprocess:     Preprocess,
	PreprocessCode: PreprocessCode,
	Normalize:      Normalize,
	Annotations:    Native,
	Code:           Code,
}
