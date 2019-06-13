// Package discovery package implements helpers for clients to discover language drivers supported by Babelfish.
//
// The drivers discovery process uses Github API under the hood to get the most up-to-date information. This process may
// fail in case of Github rate limiting, and the discovery will fallback to a drivers list hosted by Babelfish project,
// which must be reasonably up-to-date.
//
// It is also possible to provide Github API token by setting the GITHUB_TOKEN environment variable to prevent the
// rate limiting.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/bblfsh/sdk/v3/driver/manifest"
)

const (
	GithubOrg = "bblfsh"
)

var gh struct {
	once   sync.Once
	client *github.Client
}

// githubClient returns a lazily-initialized singleton Github API client.
//
// API token can be provided with GITHUB_TOKEN environment variable. The public API rate limits will apply without it.
func githubClient() *github.Client {
	gh.once.Do(func() {
		hc := http.DefaultClient
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			hc = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			))
		}
		gh.client = github.NewClient(hc)
	})
	return gh.client
}

// topics that each driver repository on Github should be annotated with
var topics = []string{
	"babelfish", "driver",
}

// Driver is an object describing language driver and it's repository-related information.
type Driver struct {
	manifest.Manifest
	repo *github.Repository
}

type byStatusAndName []Driver

func (arr byStatusAndName) Len() int {
	return len(arr)
}

func (arr byStatusAndName) Less(i, j int) bool {
	a, b := arr[i], arr[j]
	// sort by status, features count, name
	if s1, s2 := a.Status.Rank(), b.Status.Rank(); s1 > s2 {
		return true
	} else if s1 < s2 {
		return false
	}
	if n1, n2 := len(a.Features), len(b.Features); n1 > n2 {
		return true
	} else if n1 < n2 {
		return false
	}
	return a.Language < b.Language
}

func (arr byStatusAndName) Swap(i, j int) {
	arr[i], arr[j] = arr[j], arr[i]
}

// RepositoryURL returns Github repository URL for browsers (not git).
func (d Driver) RepositoryURL() string {
	return d.repo.GetHTMLURL()
}

// repositoryFileURL returns an URL of file in the driver's repository.
func (d Driver) repositoryFileURL(path string) string {
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/master/%s", d.repo.GetFullName(), path)
}

// newReq constructs a GET request with context.
func newReq(ctx context.Context, url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// if it fails, it's a programmer's error
		panic(err)
	}
	return req.WithContext(ctx)
}

// loadManifest reads manifest file from repository and decodes it into object.
func (d *Driver) loadManifest(ctx context.Context) error {
	req := newReq(ctx, d.repositoryFileURL(manifest.Filename))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// outdated driver
		d.Name = d.Language
		d.Status = manifest.Inactive
		return nil
	} else if resp.StatusCode/100 != 2 { // 2xx
		return fmt.Errorf("status: %v", resp.Status)
	}

	lang := d.Language
	if err := d.Manifest.Decode(resp.Body); err != nil {
		return err
	}
	// override language ID from manifest (prevents copy-paste of manifests)
	d.Language = lang
	if d.Name == "" {
		d.Name = d.Language
	}
	return nil
}

// fetchFromGithub returns a manifest.ReadFunc that is bound to context and fetches files directly from Github.
func (d Driver) fetchFromGithub(ctx context.Context) manifest.ReadFunc {
	return func(path string) ([]byte, error) {
		req := newReq(ctx, d.repositoryFileURL(path))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil
		} else if resp.StatusCode/100 != 2 {
			return nil, fmt.Errorf("status: %v", resp.Status)
		}
		return ioutil.ReadAll(resp.Body)
	}
}

// loadMaintainers reads MAINTAINERS file from repository and decodes it into object.
func (d *Driver) loadMaintainers(ctx context.Context) error {
	list, err := manifest.Maintainers(d.fetchFromGithub(ctx))
	if err != nil {
		return err
	}
	d.Maintainers = list
	return nil
}

// loadSDKVersion reads SDK version from repository and decodes it into object.
func (d *Driver) loadSDKVersion(ctx context.Context) error {
	vers, err := manifest.SDKVersion(d.fetchFromGithub(ctx))
	if err != nil {
		return err
	}
	d.SDKVersion = vers
	return nil
}

// loadBuildInfo build-related information from repository and decodes it into object.
func (d *Driver) loadBuildInfo(ctx context.Context, m *manifest.Manifest) error {
	return manifest.LoadRuntimeInfo(m, d.fetchFromGithub(ctx))
}

// repoName returns a user/org name and the repository name for a driver.
func (d *Driver) repoName() (string, string) {
	repo := d.repo.GetName()
	owner := d.repo.GetOrganization().GetName()
	if owner == "" {
		owner = d.repo.GetOwner().GetLogin()
	}
	return owner, repo
}

// eachTag calls a given function for each tag in the driver repository.
func (d *Driver) eachTag(ctx context.Context, fnc func(tag *github.RepositoryTag) (bool, error)) error {
	owner, repo := d.repoName()
	cli := githubClient()

	for page := 1; ; page++ {
		resp, _, err := cli.Repositories.ListTags(ctx, owner, repo, &github.ListOptions{
			Page: page, PerPage: 100,
		})
		if err != nil {
			return err
		} else if len(resp) == 0 {
			break
		}
		for _, r := range resp {
			next, err := fnc(r)
			if err != nil {
				return err
			} else if !next {
				return nil
			}
		}
	}
	return nil
}

// eachVersion calls a given function for each version in the driver repository.
func (d *Driver) eachVersion(ctx context.Context, fnc func(vers Version, tag *github.RepositoryTag) (bool, error)) error {
	return d.eachTag(ctx, func(tag *github.RepositoryTag) (bool, error) {
		name := tag.GetName()
		if name == "" || name[0] != 'v' {
			return true, nil // skip
		}
		vers, err := semver.Parse(name[1:])
		if err != nil {
			return true, nil // ignore semver parsing errors
		}
		return fnc(vers, tag)
	})
}

// Version is a semantic version.
type Version = semver.Version

// LatestVersion returns a latest version of the driver. It returns an empty version if there are no releases of the driver.
func (d *Driver) LatestVersion(ctx context.Context) (Version, error) {
	var latest Version
	err := d.eachVersion(ctx, func(v Version, _ *github.RepositoryTag) (bool, error) {
		if v.GT(latest) {
			latest = v
		}
		return true, nil
	})
	return latest, err
}

// Versions returns a list of available driver versions sorted in descending order (first element is the latest version).
func (d *Driver) Versions(ctx context.Context) ([]Version, error) {
	var vers []Version
	err := d.eachVersion(ctx, func(v Version, _ *github.RepositoryTag) (bool, error) {
		vers = append(vers, v)
		return true, nil
	})
	sort.Sort(sort.Reverse(semver.Versions(vers)))
	return vers, err
}

// Options controls how drivers are being discovered and what information is fetched for them.
type Options struct {
	Organization  string // Github organization name
	NamesOnly     bool   // driver manifest will only have Language field populated
	NoMaintainers bool   // do not load maintainers list
	NoSDKVersion  bool   // do not check SDK version
	NoBuildInfo   bool   // do not load build info
	NoStatic      bool   // do not use a static manifest - discover drivers
}

// isRateLimit checks if error is due to rate limiting.
func isRateLimit(err error) bool {
	_, ok := err.(*github.RateLimitError)
	return ok
}

// getDriversForOrg lists all repositories for an organization and filters ones that contains topics of the driver.
func getDriversForOrg(ctx context.Context, org string) ([]Driver, error) {
	cli := githubClient()

	var out []Driver
	// list all repositories in organization
	for page := 1; ; page++ {
		repos, _, err := cli.Repositories.ListByOrg(ctx, org, &github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{
				Page: page, PerPage: 100,
			},
			Type: "public",
		})
		if err != nil {
			return out, err
		} else if len(repos) == 0 {
			break
		}
		for _, r := range repos {
			// filter repos by topics to find babelfish drivers
			if containsTopics(r.Topics, topics...) {
				out = append(out, Driver{
					Manifest: manifest.Manifest{
						Language: strings.TrimSuffix(r.GetName(), "-driver"),
					},
					repo: r,
				})
			}
		}
	}
	return out, nil
}

const staticDriversURL = `https://raw.githubusercontent.com/` + GithubOrg + `/documentation/master/languages.json`

// getStaticDrivers downloads a static drivers list hosted by Babelfish org.
func getStaticDrivers(ctx context.Context) ([]Driver, error) {
	req, err := http.NewRequest("GET", staticDriversURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot download static driver list: status: %v", resp.Status)
	}
	var drivers []Driver
	err = json.NewDecoder(resp.Body).Decode(&drivers)
	if err != nil {
		return nil, fmt.Errorf("cannot decode static driver list: %v", err)
	}
	return drivers, nil
}

// OfficialDrivers lists all available language drivers for Babelfish.
func OfficialDrivers(ctx context.Context, opt *Options) ([]Driver, error) {
	if opt == nil {
		opt = &Options{}
	}
	if opt.Organization == "" {
		opt.Organization = GithubOrg
	}
	out, err := getDriversForOrg(ctx, opt.Organization)
	if isRateLimit(err) && opt.Organization == GithubOrg && !opt.NoStatic {
		return getStaticDrivers(ctx)
	} else if err != nil {
		return out, err
	}
	if opt.NamesOnly {
		sort.Sort(byStatusAndName(out))
		return out, nil
	}

	// load manifest and maintainers file from repositories
	var (
		wg sync.WaitGroup
		// limits the number of concurrent requests
		tokens = make(chan struct{}, 3)

		mu   sync.Mutex
		last error
	)

	setErr := func(err error) {
		mu.Lock()
		last = err
		mu.Unlock()
	}
	for i := range out {
		wg.Add(1)
		go func(d *Driver) {
			defer wg.Done()

			tokens <- struct{}{}
			defer func() {
				<-tokens
			}()
			if err := d.loadManifest(ctx); err != nil {
				setErr(err)
			}
			if !opt.NoSDKVersion {
				if err := d.loadSDKVersion(ctx); err != nil {
					setErr(err)
				}
			}
			if !opt.NoBuildInfo {
				if err := d.loadBuildInfo(ctx, &d.Manifest); err != nil {
					setErr(err)
				}
			}
			if !opt.NoMaintainers {
				if err := d.loadMaintainers(ctx); err != nil {
					setErr(err)
				}
			}
		}(&out[i])
	}
	wg.Wait()
	sort.Sort(byStatusAndName(out))
	return out, last
}

// containsTopics returns true if all inc topics are present in the list.
func containsTopics(topics []string, inc ...string) bool {
	n := 0
	for _, t := range topics {
		ok := false
		for _, t2 := range inc {
			if t == t2 {
				ok = true
				break
			}
		}
		if ok {
			n++
		}
	}
	return n == len(inc)
}
