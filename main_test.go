package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestIsValidMetricType(t *testing.T) {
	var typesTest = []struct {
		input string
		want  bool
	}{
		{input: "", want: true},
		{input: "gauge", want: true},
		{input: "rate", want: true},
		{input: "count", want: true},
		{input: "Foo", want: false},
	}

	for _, tt := range typesTest {
		if got := isValidMetricType(tt.input); got != tt.want {
			t.Fatalf("expected %v, got %v", tt.want, got)
		}
	}
}

func TestIsValidAlertType(t *testing.T) {
	var typesTest = []struct {
		input string
		want  bool
	}{
		{input: "", want: true},
		{input: "info", want: true},
		{input: "success", want: true},
		{input: "warning", want: true},
		{input: "error", want: true},
		{input: "Foo", want: false},
	}

	for _, tt := range typesTest {
		if got := isValidAlertType(tt.input); got != tt.want {
			t.Fatalf("expected %v, got %v", tt.want, got)
		}
	}
}

func TestIsValidPriority(t *testing.T) {
	var typesTest = []struct {
		input string
		want  bool
	}{
		{input: "", want: true},
		{input: "low", want: true},
		{input: "normal", want: true},
		{input: "Foo", want: false},
	}

	for _, tt := range typesTest {
		if got := isValidPriority(tt.input); got != tt.want {
			t.Fatalf("expected %v, got %v", tt.want, got)
		}
	}
}

func TestIsValidAggregationKey(t *testing.T) {
	var typesTest = []struct {
		input string
		want  bool
	}{
		{input: "", want: true},
		{input: "this_is_a_good_key", want: true},
		{input: strings.Repeat("a", 101), want: false},
	}

	for _, tt := range typesTest {
		if got := isValidAggregationKey(tt.input); got != tt.want {
			t.Fatalf("expected %v, got %v", tt.want, got)
		}
	}
}

func TestPrintVersion(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	log.SetFlags(0)

	printVersion()
	want := "Drone-Datadog Plugin version: snapshot - git commit: -"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("expected '%v', got '%v'", want, got)
	}

	buf.Reset()
	BuildCommit = "abdcdef"
	BuildTag = "v0.4.2"
	printVersion()
	want = "Drone-Datadog Plugin version: v0.4.2 - git commit: abdcdef"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("expected '%v', got '%v'", want, got)
	}
}

func TestParseConfig(t *testing.T) {
	// no dry run and no api key should raise an error
	os.Setenv("PLUGIN_DRY_RUN", "false")
	if _, err := parseConfig(); err == nil {
		t.Fatalf("parseConfig should raise an error")
	}

	want := "123456"
	os.Setenv("PLUGIN_API_KEY", want)
	cfg, err := parseConfig()
	if err != nil {
		t.Fatal("expected no error")
	}
	if cfg.APIKey != want {
		t.Fatalf("expected '%v', got '%v'", want, cfg.APIKey)
	}
}

func TestCustomMarshalling(t *testing.T) {
	m := Metric{
		Name:  "foo.bar.baz",
		Type:  "count",
		Value: 1.0,
		Host:  "127.0.0.1",
		Tags:  []string{"foo:bar", "foo:baz"},
	}

	// trim the part containing the timestamp
	want := `{"metric":"foo.bar.baz","type":"count","host":"127.0.0.1","tags":["foo:bar","foo:baz"],"points":[[`
	got, err := m.MarshalJSON()
	if err != nil {
		t.Fatal("expected no error")
	}
	if !strings.Contains(string(got), want) {
		t.Fatalf("expected '%v', got '%v'", want, string(got))
	}
}

func TestSend(t *testing.T) {
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	log.SetFlags(0)

	// when dry run is true, output should be go to the logger
	send("http://example.com", []byte("Hello, World!"), true)

	want := "Dry run, logging payload:\nHello, World!"
	if got := strings.TrimSpace(buf.String()); got != want {
		t.Fatalf("expected '%v', got '%v'", want, got)
	}

	// let it make the actual HTTP call
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ := ioutil.ReadAll(r.Body)
		if string(got) != "Hello, World!" {
			t.Fatalf("expected '%v', got '%v'", "Hello, World!", string(got))
		}
		if r.Method != "POST" {
			t.Fatalf("expected '%v', got '%v'", "POST", r.Method)
		}
		if val := r.Header.Get("Content-Type"); val != "application/json" {
			t.Fatalf("wrong Content-Type header: %v", val)
		}
	}))
	defer ts.Close()

	if err := send(ts.URL, []byte("Hello, World!"), false); err != nil {
		t.Fatal("expected no error")
	}
}
