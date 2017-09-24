package manifest

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

const Filename = "manifest.toml"

type DevelopmentStatus string

const (
	Planning DevelopmentStatus = "planning"
	PreAlpha DevelopmentStatus = "pre-alpha"
	Alpha    DevelopmentStatus = "alpha"
	Beta     DevelopmentStatus = "beta"
	Stable   DevelopmentStatus = "stable"
	Mature   DevelopmentStatus = "mature"
	Inactive DevelopmentStatus = "inactive"
)

// InformationLoss in terms of which kind of code generation would they allow.
type InformationLoss string

const (
	// Lossless no information loss converting code to AST and then back to code
	// would. code == codegen(AST(code)).
	Lossless InformationLoss = "lossless"
	// FormatingLoss only superfluous formatting information is lost (e.g.
	// whitespace, indentation). Code generated from the AST could be the same
	// as the original code after passing a code formatter.
	// fmt(code) == codegen(AST(code)).
	FormatingLoss InformationLoss = "formating-loss"
	// SyntacticSugarLoss there is information loss about syntactic sugar. Code
	// generated from the AST could be the same as the original code after
	// desugaring it. desugar(code) == codegen(AST(code)).
	SyntacticSugarLoss InformationLoss = "syntactic-sugar-loss"
	// CommentLoss comments are not present in the AST.
	CommentLoss InformationLoss = "formating-loss"
)

type OS string

const (
	Alpine OS = "alpine"
	Debian OS = "debian"
)

func (os OS) AsImage() string {
	switch os {
	case Alpine:
		return "alpine:3.6"
	case Debian:
		return "debian:jessie-slim"
	default:
		return ""
	}
}

type Manifest struct {
	Language        string            `toml:"language"`
	Version         string            `toml:"version,omitempty"`
	Build           *time.Time        `toml:"build,omitempty"`
	Status          DevelopmentStatus `toml:"status"`
	InformationLoss []InformationLoss `toml:"loss"`
	Documentation   struct {
		Description string `toml:"description,omitempty"`
		Caveats     string `toml:"caveats,omitempty"`
	} `toml:"documentation,omitempty"`
	Runtime struct {
		OS            OS       `toml:"os"`
		NativeVersion Versions `toml:"native_version"`
		GoVersion     string   `toml:"go_version"`
	} `toml:"runtime"`
}

type Versions []string

func (v Versions) String() string {
	return strings.Join(v, ":")
}

// Load reads a manifest and decode the content into a new Manifest struct
func Load(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	m := &Manifest{}
	return m, m.Decode(f)
}

// Encode encodes m in toml format and writes the restult to w
func (m *Manifest) Encode(w io.Writer) error {
	e := toml.NewEncoder(w)
	return e.Encode(m)
}

// Decode decodes reads r and decodes it into m
func (m *Manifest) Decode(r io.Reader) error {
	if _, err := toml.DecodeReader(r, m); err != nil {
		return err
	}

	return nil
}
