package protocol

import (
	"context"
	"errors"
	"net"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	serrors "gopkg.in/src-d/go-errors.v1"

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

var _ driver.Driver = (*driverMock)(nil)

type driverMock struct {
	name string
	uast nodes.Node
	vers driver.Version
	list []manifest.Manifest
	err  error
}

func (d *driverMock) Parse(ctx context.Context, src string, opts *driver.ParseOptions) (nodes.Node, error) {
	return d.uast, d.err
}

func (d *driverMock) Version(ctx context.Context) (driver.Version, error) {
	return d.vers, d.err
}

func (d *driverMock) Languages(ctx context.Context) ([]manifest.Manifest, error) {
	return d.list, d.err
}

func defaultUAST() nodes.Node {
	return nodes.Object{"k": nodes.String("v")}
}

func kindOfError(e *serrors.Error) *serrors.Kind {
	// FIXME(dennwc): fix upstream
	return (*struct {
		kind *serrors.Kind
	})(unsafe.Pointer(e)).kind
}

func equalError(t testing.TB, exp, got error) {
	if exp, ok := exp.(*serrors.Error); ok {
		if got, ok := got.(*serrors.Error); ok {
			kexp, kgot := kindOfError(exp), kindOfError(got)
			if kexp != kgot {
				require.Fail(t, "unexpected error kind", "%v vs %v", kexp, kgot)
			}
			equalError(t, exp.Cause(), got.Cause())
			return
		}
	}
	require.Equal(t, exp, got)
}

func TestDriverError(t *testing.T) {
	var cases = []driverMock{
		{name: "success", uast: defaultUAST()},
		{name: "partial parse", uast: defaultUAST(), err: driver.ErrSyntax.Wrap(errors.New("invalid source"))},
		{name: "non-utf8", err: driver.ErrUnknownEncoding.New()},
		{name: "language detection failure", err: driver.ErrLanguageDetection.New()},
		{name: "unsupported mode", err: driver.ErrModeNotSupported.New()},
		{name: "transform failure", err: driver.ErrTransformFailure.Wrap(errors.New("test failure"))},
		{name: "driver failure", err: driver.ErrDriverFailure.Wrap(errors.New("test failure"))},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var d driver.Driver = &c
			srv := grpc.NewServer(ServerOptions()...)
			RegisterDriver(srv, d)
			errc := make(chan error, 1)
			lis, err := net.Listen("tcp", "localhost:0")
			require.NoError(t, err)
			defer lis.Close()
			go func() {
				if err := srv.Serve(lis); err != nil {
					errc <- err
				}
			}()

			opts := append([]grpc.DialOption{grpc.WithInsecure()}, DialOptions()...)
			cc, err := grpc.Dial(lis.Addr().String(), opts...)
			if err != nil {
				select {
				case serr := <-errc:
					err = serr
				default:
				}
			}
			require.NoError(t, err)
			defer cc.Close()

			cd := AsDriver(cc)

			nd, err := cd.Parse(context.Background(), "test", nil)
			exp, eerr := d.Parse(context.Background(), "test", nil)
			equalError(t, eerr, err)
			require.Equal(t, exp, nd)
		})
	}
}
