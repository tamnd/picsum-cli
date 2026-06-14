package picsum_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tamnd/picsum-cli/picsum"
)

const fakeListJSON = `[
  {"id":"0","author":"Alejandro Escamilla","width":5616,"height":3744,"url":"https://unsplash.com/photos/yC-Yzbqy7PY","download_url":"https://picsum.photos/id/0/5616/3744"},
  {"id":"1","author":"Alejandro Escamilla","width":5616,"height":3744,"url":"https://unsplash.com/photos/R_Rs0oQmZNk","download_url":"https://picsum.photos/id/1/5616/3744"}
]`

const fakeInfoJSON = `{"id":"42","author":"Alejandro Escamilla","width":5000,"height":3333,"url":"https://unsplash.com/photos/N7XodRrbzS0","download_url":"https://picsum.photos/id/42/5000/3333"}`

func newTestClient(ts *httptest.Server) *picsum.Client {
	cfg := picsum.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return picsum.NewClient(cfg)
}

func TestListParsesImages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeListJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	images, err := c.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 2 {
		t.Fatalf("got %d images, want 2", len(images))
	}
	if images[0].ID != "0" {
		t.Errorf("images[0].ID = %q, want 0", images[0].ID)
	}
	if images[0].Author != "Alejandro Escamilla" {
		t.Errorf("images[0].Author = %q, want Alejandro Escamilla", images[0].Author)
	}
	if images[1].ID != "1" {
		t.Errorf("images[1].ID = %q, want 1", images[1].ID)
	}
}

func TestListHitsCorrectPath(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		_, _ = fmt.Fprint(w, fakeListJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.List(context.Background(), 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v2/list?page=2&limit=5" {
		t.Errorf("path = %q, want /v2/list?page=2&limit=5", gotPath)
	}
}

func TestInfoParsesImage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeInfoJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	img, err := c.Info(context.Background(), "42")
	if err != nil {
		t.Fatal(err)
	}
	if img.ID != "42" {
		t.Errorf("ID = %q, want 42", img.ID)
	}
	if img.Author != "Alejandro Escamilla" {
		t.Errorf("Author = %q, want Alejandro Escamilla", img.Author)
	}
	if img.Width != 5000 {
		t.Errorf("Width = %d, want 5000", img.Width)
	}
	if img.Height != 3333 {
		t.Errorf("Height = %d, want 3333", img.Height)
	}
}

func TestInfoHitsCorrectPath(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = fmt.Fprint(w, fakeInfoJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Info(context.Background(), "42")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/id/42/info" {
		t.Errorf("path = %q, want /id/42/info", gotPath)
	}
}

func TestListSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeListJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestListRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeListJSON)
	}))
	defer ts.Close()

	cfg := picsum.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 4
	c := picsum.NewClient(cfg)

	_, err := c.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}
