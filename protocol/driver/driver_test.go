package driver

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v1/protocol"

	"github.com/stretchr/testify/require"
)

func TestDriverParser(t *testing.T) {
	require := require.New(t)

	d := &Driver{}
	d.Binary = "internal/native/mock"

	err := d.Start()
	require.NoError(err)

	r := d.Parse(&protocol.ParseRequest{
		Content: "foo",
	})

	require.NotNil(r)

	err = d.Stop()
	require.NoError(err)
	require.Equal(r.Status, protocol.Ok)
	require.Equal(r.UAST.String(), " "+
		"{\n"+
		".  Roles: Unannotated\n"+
		".  Properties: {\n"+
		".  .  key: val\n"+
		".  }\n"+
		"}\n",
	)
}

func TestDriverParser_Malfunctioning(t *testing.T) {
	require := require.New(t)

	d := &Driver{}
	d.Binary = "echo"

	err := d.Start()
	require.NoError(err)

	r := d.Parse(&protocol.ParseRequest{
		Content: "foo",
	})

	require.NotNil(r)

	require.Equal(r.Status, protocol.Fatal)
	require.Equal(len(r.Errors), 1)
	require.Nil(r.UAST)

	err = d.Stop()
	require.NoError(err)
}
