package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Job struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	CurrentStep string `json:"current_step"`
}

type Client struct {
	base  string
	token string
	http  *http.Client
}

func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" { return v }
	return d
}

func (c *Client) request(method, path string, body any) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil { return nil, err }
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.base+path, reader)
	if err != nil { return nil, err }
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil { req.Header.Set("Content-Type", "application/json") }
	return c.http.Do(req)
}

func (c *Client) next() (*Job, error) {
	resp, err := c.request(http.MethodGet, "/api/worker/jobs/next", nil)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent { return nil, nil }
	if resp.StatusCode != http.StatusOK { return nil, errors.New(resp.Status) }
	var job Job
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil { return nil, err }
	return &job, nil
}

func (c *Client) patch(id string, payload any) error {
	resp, err := c.request(http.MethodPatch, "/api/worker/deployments/"+id, payload)
	if err != nil { return err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return errors.New(resp.Status) }
	return nil
}

func main() {
	api := env("REFORGE_API_URL", "http://api:8080")
	token := env("REFORGE_WORKER_TOKEN", "")
	if len(token) < 32 {
		log.Fatal("REFORGE_WORKER_TOKEN must be configured with at least 32 characters")
	}
	execute := strings.EqualFold(env("REFORGE_EXECUTE_PRIVILEGED", "false"), "true")
	client := &Client{base: api, token: token, http: &http.Client{Timeout: 15 * time.Second}}

	for {
		job, err := client.next()
		if err != nil {
			log.Printf("queue error: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if job == nil {
			time.Sleep(2 * time.Second)
			continue
		}
		if !execute {
			_ = client.patch(job.ID, map[string]any{
				"status": "waiting", "progress": 0, "current_step": "Waiting for imaging node",
			})
			continue
		}

		// Privileged raw-disk execution belongs to the separately installed imaging node.
		// The orchestration worker must not pretend that an image was applied.
		_ = client.patch(job.ID, map[string]any{
			"status": "waiting", "progress": 0, "current_step": "Imaging node required",
		})
	}
}
