package kc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rendau/jwts/internal/service/jwk/model"
)

type Service struct {
	http      *http.Client
	url       string
	realmName string
}

func New(url, realmName string) *Service {
	if url == "" || realmName == "" {
		return nil
	}

	return &Service{
		http: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost: 100,
			},
		},
		url:       url,
		realmName: realmName,
	}
}

func (s *Service) FetchJwks(ctx context.Context) (*model.JwkSet, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", s.url, s.realmName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("kc.service - fetch jwks - build request: %w", err)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kc.service - fetch jwks - do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("kc.service - fetch jwks - bad status %s: %s", resp.Status, string(b))
	}

	var result *model.JwkSet
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("kc.service - fetch jwks - decode jwks: %w", err)
	}

	return result, nil
}
