package discovery

import (
	"context"
	"testing"

	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/blang/semver"
	"github.com/stretchr/testify/require"
)

func TestOfficialDrivers(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := context.Background()
	drivers, err := OfficialDrivers(ctx, nil)
	if isRateLimit(err) {
		t.Skip(err)
	}
	require.NoError(t, err)
	require.True(t, len(drivers) >= 15, "drivers: %d", len(drivers))

	// make sure that IDs are distinct
	m := make(map[string]Driver)
	for _, d := range drivers {
		m[d.Language] = d
	}

	for _, exp := range []Driver{
		{Manifest: manifest.Manifest{Language: "go", Name: "Go"}},
		{Manifest: manifest.Manifest{Language: "javascript", Name: "JavaScript"}},
	} {
		got := m[exp.Language]
		require.Equal(t, exp.Language, got.Language)
		require.Equal(t, exp.Name, got.Name)
		require.NotEmpty(t, got.Maintainers)
		require.NotEmpty(t, got.Features)

		if exp.Language == "go" {
			const latest = "2.7.1"
			vers, err := got.Versions(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, vers)
			require.True(t, len(vers) >= 18, "versions: %d", len(vers))
			require.True(t, semver.MustParse(latest).GTE(vers[0]))
		}
	}
}
