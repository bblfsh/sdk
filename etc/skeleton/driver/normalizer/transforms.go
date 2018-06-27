package normalizer

import "gopkg.in/bblfsh/sdk.v2/sdk/driver"

var Transforms = driver.Transforms{
	Preprocess: Preprocess,
	Normalize:  Normalize,
	Native:     Native,
	Code:       Code,
}
