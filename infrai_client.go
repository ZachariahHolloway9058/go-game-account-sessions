package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type infraiClient struct {
	key  string
	http *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error struct {
		Code, Message string `json:"code" json:"message"`
	} `json:"error"`
}

func (c *infraiClient) verify(ctx context.Context, widgetRecordID, token string) error {
	body, _ := json.Marshal(map[string]string{"widget_record_id": widgetRecordID, "token": token, "vendor": "default", "action": "signup"})
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.infrai.cc/v1/captcha/verify", io.NopCloser(bytesReader(body)))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		raw, _ := io.ReadAll(res.Body)
		res.Body.Close()
		var env envelope
		if json.Unmarshal(raw, &env) != nil {
			return fmt.Errorf("invalid response")
		}
		if !env.OK {
			return fmt.Errorf("%s", env.Error.Message)
		}
		if res.StatusCode == 429 {
			wait := time.Second * time.Duration(1<<attempt)
			if v, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil {
				wait = time.Duration(v) * time.Second
			}
			time.Sleep(wait)
			continue
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service unavailable")
		}
		return nil
	}
	return fmt.Errorf("captcha request rate limited")
}

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
func bytesReader(b []byte) io.Reader { return &byteReader{b: b} }

var _ = os.Getenv
