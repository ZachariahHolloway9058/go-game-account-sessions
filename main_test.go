package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCaptchaVerifyFields(t *testing.T) {
	c := &infraiClient{key: "test", http: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["widget_record_id"] != "widget-123" || body["token"] != "token-123" {
			t.Fatalf("unexpected captcha request: %v", body)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"ok":true}`)), Header: make(http.Header)}, nil
	})}}
	if err := c.verify(context.Background(), "widget-123", "token-123"); err != nil {
		t.Fatal(err)
	}
}

func TestSignupAndLogin(t *testing.T) {
	s := &service{players: map[string]player{}, sessions: map[string]string{}, captcha: &infraiClient{}}
	cases := []struct{ name, path, body, want string }{{"signup", "/signup", `{"email":"a@b.com","password":"pw","name":"A"}`, "registered"}, {"login", "/login", `{"email":"a@b.com","password":"pw"}`, "session_id"}}
	for _, tc := range cases {
		r := httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		if tc.path == "/signup" {
			s.signup(w, r)
		} else {
			s.login(w, r)
		}
		if !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("%s: %s", tc.name, w.Body.String())
		}
	}
}
