package main

import (
	"github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer"

	"gopkg.in/bblfsh/sdk.v1/sdk/driver"
)

func main() {
	d, err := driver.NewDriver(normalizer.ToNode, normalizer.Transformers)
	if err != nil {
		panic(err)
	}

	s := driver.NewServer(d)
	if err := s.Start(); err != nil {
		panic(err)
	}
}
