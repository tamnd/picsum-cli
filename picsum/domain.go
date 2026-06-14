package picsum

import (
	"context"
	"fmt"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes Lorem Picsum as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/picsum-cli/picsum"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// picsum:// URIs by routing to the operations Register installs. The standalone
// picsum binary does not use any of this, so the CLI is unchanged.
func init() { kit.Register(Domain{}) }

// Domain is the picsum driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against, and
// the identity a host reuses for help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "picsum",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "picsum",
			Short:  "Browse Lorem Picsum placeholder images",
			Long: `picsum fetches image metadata from the Lorem Picsum public API.
No API key required.`,
			Site: Host,
			Repo: "https://github.com/tamnd/picsum-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// list: paginated catalogue of available images
	kit.Handle(app, kit.OpMeta{
		Name:    "list",
		Group:   "read",
		Summary: "List available images",
	}, listOp)

	// image: fetch one image's metadata by numeric id
	kit.Handle(app, kit.OpMeta{
		Name:    "image",
		Group:   "read",
		Single:  true,
		Summary: "Get info about a specific image by ID",
		URIType: "image",
		Resolver: true,
		Args:    []kit.Arg{{Name: "id", Help: "image id (e.g. 42)"}},
	}, imageOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type listInput struct {
	Page   int     `kit:"flag" help:"page number" default:"1"`
	Limit  int     `kit:"flag" help:"images per page" default:"20"`
	Client *Client `kit:"inject"`
}

type imageInput struct {
	ID     string  `kit:"arg" help:"image id (e.g. 42)"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func listOp(ctx context.Context, in listInput, emit func(Image) error) error {
	page := in.Page
	if page < 1 {
		page = 1
	}
	limit := in.Limit
	if limit < 1 {
		limit = 20
	}
	images, err := in.Client.List(ctx, page, limit)
	if err != nil {
		return mapErr(err)
	}
	for _, img := range images {
		if err := emit(img); err != nil {
			return err
		}
	}
	return nil
}

func imageOp(ctx context.Context, in imageInput, emit func(Image) error) error {
	img, err := in.Client.Info(ctx, in.ID)
	if err != nil {
		return mapErr(err)
	}
	return emit(img)
}

// --- Resolver ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("empty picsum reference")
	}
	return "image", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "image":
		return fmt.Sprintf("https://picsum.photos/id/%s/info", id), nil
	default:
		return "", errs.Usage("picsum has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
