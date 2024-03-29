package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	bimg "gopkg.in/h2non/bimg.v1"
)

func TestIndex(t *testing.T) {
	ts := testServer(indexController)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: %s", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(body), "thumbnary") == false {
		t.Fatalf("Invalid body response: %s", body)
	}
}

func TestCrop(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	url := ts.URL + "/c!/w=300/testdata/large.jpg?origin=qic0bfzg"
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	if res.Header.Get("Content-Length") == "" {
		t.Fatal("Empty content length response")
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 1080)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func TestFit(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	url := ts.URL + "/c!/w=300,h=300,m=fit/testdata/large.jpg?origin=qic0bfzg"
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 168)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func TestTypeAuto(t *testing.T) {
	cases := []struct {
		acceptHeader string
		expected     string
	}{
		{"", "jpeg"},
		{"image/webp,*/*", "webp"},
		{"image/png,*/*", "png"},
		{"image/webp;q=0.8,image/jpeg", "webp"},
		{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8", "webp"}, // Chrome
	}

	for _, test := range cases {
		opts := ServerOptions{
			OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
		}
		opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			buf, _ := ioutil.ReadFile("testdata/large.jpg")
			w.Write(buf)
		}))
		defer td()

		fn := ImageMiddleware(opts)
		ts := httptest.NewServer(fn)
		url := ts.URL + "/c!/w=300,f=auto/testdata/large.jpg?origin=qic0bfzg"
		defer ts.Close()

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("Content-Type", "image/jpeg")
		req.Header.Add("Accept", test.acceptHeader)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal("Cannot perform the request")
		}
		if res.StatusCode != 200 {
			t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
		}

		if res.Header.Get("Content-Length") == "" {
			t.Fatal("Empty content length response")
		}

		image, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if len(image) == 0 {
			t.Fatalf("Empty response body")
		}

		err = assertSize(image, 300, 1080)
		if err != nil {
			t.Error(err)
		}

		if bimg.DetermineImageTypeName(image) != test.expected {
			t.Fatalf("Image type expected to '%s', but actual '%s'. (req=%+v)", test.expected, bimg.DetermineImageTypeName(image), req)
		}

		if res.Header.Get("Vary") != "Accept" {
			t.Fatal("Vary header not set correctly")
		}
	}
}

func setupTestSourceServer(opts ServerOptions, httpFunc http.HandlerFunc) (ServerOptions, func()) {
	LoadSources(opts)

	tsImage := httptest.NewServer(httpFunc)

	tsImageURL, _ := url.Parse(tsImage.URL)

	originMap := map[OriginSlug]*Origin{
		"qic0bfzg": &Origin{
			Slug:       "qic0bfzg",
			SourceType: ImageSourceTypeHttp,
			Scheme:     tsImageURL.Scheme,
			Host:       tsImageURL.Host,
			PathPrefix: "/",
		},
		"jdv9ab8v": &Origin{
			Slug:                    "jdv9ab8v",
			SourceType:              ImageSourceTypeHttp,
			Scheme:                  tsImageURL.Scheme,
			Host:                    tsImageURL.Host,
			PathPrefix:              "/",
			URLSignatureEnabled:     true,
			URLSignatureKey:         "secrettest",
			URLSignatureKey_Version: 1,
		},
		"sigver2": &Origin{
			Slug:                     "sigver2",
			SourceType:               ImageSourceTypeHttp,
			Scheme:                   tsImageURL.Scheme,
			Host:                     tsImageURL.Host,
			PathPrefix:               "/",
			URLSignatureEnabled:      true,
			URLSignatureKey:          "secrettestnew",
			URLSignatureKey_Previous: "secrettest",
			URLSignatureKey_Version:  2,
		},
	}
	opts.OriginRepos = NewMockOriginRepository(originMap)

	return opts, func() {
		tsImage.Close()
	}
}

func TestRemoteHTTPSource(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	url := ts.URL + "/c!/w=200,h=200/testdata/large.jpg?origin=qic0bfzg"
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 200, 200)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func TestInvalidRemoteHTTPSource(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(400)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	LoadSources(opts)

	ts := httptest.NewServer(fn)
	url := ts.URL + "/c!/w=200,h=200/testdata/large.jpg?origin=qic0bfzg"
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Request failed")
	}
	if res.StatusCode != 400 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}
}

func TestURLSignature(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	path := "/c!/w=300/testdata/large.jpg"
	sig := CreateURLSignatureString(1, path, "secrettest", "")
	url := ts.URL + path + "?origin=jdv9ab8v" + "&sig=" + sig
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	if res.Header.Get("Content-Length") == "" {
		t.Fatal("Empty content length response")
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 1080)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func TestURLSignatureVersion2(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	path := "/c!/w=300/testdata/large.jpg"
	sig := CreateURLSignatureString(2, path, "secrettestnew", "")
	url := ts.URL + path + "?origin=sigver2" + "&sig=" + sig
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	if res.Header.Get("Content-Length") == "" {
		t.Fatal("Empty content length response")
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 1080)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func TestURLSignatureVersion2Previous(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	path := "/c!/w=300/testdata/large.jpg"
	sig := CreateURLSignatureString(1, path, "secrettest", "")
	url := ts.URL + path + "?origin=sigver2" + "&sig=" + sig
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	if res.Header.Get("Content-Length") == "" {
		t.Fatal("Empty content length response")
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 300, 1080)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func TestExternalHTTPSource(t *testing.T) {
	opts := ServerOptions{
		OriginSlugDetectMethods: []OriginSlugDetectMethod{"query"},
		AllowExternalHTTPSource: true,
	}
	opts, td := setupTestSourceServer(opts, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.String() != "/testdata/large.jpg" {
			t.Fatal("Bad request: " + req.URL.String())
		}
		buf, _ := ioutil.ReadFile("testdata/large.jpg")
		w.Write(buf)
	}))
	defer td()

	fn := ImageMiddleware(opts)
	ts := httptest.NewServer(fn)
	tsImageOrigin, _ := opts.OriginRepos.Get("jdv9ab8v")
	urlFilePath := url.QueryEscape(tsImageOrigin.Scheme + "://" + tsImageOrigin.Host + "/testdata/large.jpg")
	path := "/c!/w=200,h=200/" + urlFilePath
	sig := CreateURLSignatureString(1, path, "secrettest", "")
	url := ts.URL + path + "?origin=jdv9ab8v" + "&sig=" + sig
	defer ts.Close()

	res, err := http.Get(url)
	if err != nil {
		t.Fatal("Cannot perform the request")
	}
	if res.StatusCode != 200 {
		t.Fatalf("Invalid response status: (url=%+v) (res=%+v) (body=%s)", url, res, BodyAsString(res))
	}

	image, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(image) == 0 {
		t.Fatalf("Empty response body")
	}

	err = assertSize(image, 200, 200)
	if err != nil {
		t.Error(err)
	}

	if bimg.DetermineImageTypeName(image) != "jpeg" {
		t.Fatalf("Invalid image type")
	}
}

func testServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func readFile(file string) io.Reader {
	buf, _ := os.Open(path.Join("testdata", file))
	return buf
}

func assertSize(buf []byte, width, height int) error {
	size, err := bimg.NewImage(buf).Size()
	if err != nil {
		return err
	}
	if size.Width != width || size.Height != height {
		return fmt.Errorf("Invalid image size: exprected %dx%d, but actual %dx%d", width, height, size.Width, size.Height)
	}
	return nil
}

func BodyAsString(res *http.Response) string {
	contentLength, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	body := make([]byte, contentLength)
	len, _ := res.Body.Read(body)
	return string(body[:len])
}

type MockOriginRepository struct {
	Origins map[OriginSlug]*Origin
}

func NewMockOriginRepository(origins map[OriginSlug]*Origin) OriginRepository {
	return &MockOriginRepository{Origins: origins}
}

func (repo *MockOriginRepository) Open() error {
	return nil
}

func (repo *MockOriginRepository) Close() {
}

func (repo *MockOriginRepository) Get(originSlug OriginSlug) (*Origin, error) {
	origin, ok := repo.Origins[originSlug]
	if !ok {
		return nil, fmt.Errorf("Origin not found: (originSlug=%s)", originSlug)
	}

	return origin, nil
}
