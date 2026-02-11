package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

type ProxyServer struct {
	Providers []*Provider
	Logger    *log.Logger
	Client    *http.Client
}

func NewProxyServer(providers []*Provider, logger *log.Logger) *ProxyServer {
	return &ProxyServer{
		Providers: providers,
		Logger:    logger,
		Client: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

func (s *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadGateway)
		return
	}
	r.Body.Close()

	for _, p := range s.Providers {
		if !p.IsHealthy() {
			s.Logger.Printf("[%s] skipping (unhealthy, backoff %v)", p.Name, p.Backoff)
			continue
		}

		s.Logger.Printf("[%s] trying %s %s", p.Name, r.Method, r.URL.Path)
		resp, err := s.forwardRequest(r, p, bodyBytes)
		if err != nil {
			s.Logger.Printf("[%s] request error: %v", p.Name, err)
			p.MarkFailed()
			continue
		}

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			s.Logger.Printf("[%s] got %d, failing over", p.Name, resp.StatusCode)
			resp.Body.Close()
			p.MarkFailed()
			continue
		}

		p.MarkHealthy()
		s.Logger.Printf("[%s] success %d", p.Name, resp.StatusCode)
		s.copyResponse(w, resp)
		return
	}

	http.Error(w, "all providers failed", http.StatusBadGateway)
}

func (s *ProxyServer) forwardRequest(r *http.Request, p *Provider, body []byte) (*http.Response, error) {
	// Inject model into request body if provider has one
	modifiedBody := body
	if p.Model != "" {
		modifiedBody = injectModel(body, p.Model)
	}

	targetURL := singleJoiningSlash(p.BaseURL.String(), r.URL.Path)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, bytes.NewReader(modifiedBody))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for k, vv := range r.Header {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	// Override auth
	req.Header.Set("x-api-key", p.Token)
	req.Header.Set("Authorization", "Bearer "+p.Token)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(modifiedBody)))

	return s.Client.Do(req)
}

func (s *ProxyServer) copyResponse(w http.ResponseWriter, resp *http.Response) {
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Stream SSE responses
	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		flusher, ok := w.(http.Flusher)
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				w.Write(buf[:n])
				if ok {
					flusher.Flush()
				}
			}
			if err != nil {
				break
			}
		}
	} else {
		io.Copy(w, resp.Body)
	}
}

func injectModel(body []byte, model string) []byte {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return body
	}
	data["model"] = model
	modified, err := json.Marshal(data)
	if err != nil {
		return body
	}
	return modified
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// StartProxy starts the proxy server and returns the port.
func StartProxy(providers []*Provider, listenAddr string, logger *log.Logger) (int, error) {
	srv := NewProxyServer(providers, logger)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return 0, fmt.Errorf("listen: %w", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	go http.Serve(ln, srv)

	return port, nil
}
