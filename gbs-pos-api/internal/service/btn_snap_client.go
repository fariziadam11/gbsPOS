package service

import (
	"bytes"
	"context"
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gbs-pos-api/internal/config"
	"gbs-pos-api/internal/dto"
)

const (
	btnTokenPath    = "/snap/v1/access-token/b2b"
	btnGeneratePath = "/snap/v1/qr/qr-mpm-generate"
	btnQueryPath    = "/snap/v1/qr/qr-mpm-query"
)

type BtnSnapClient struct {
	cfg        *config.Config
	httpClient *http.Client
	tokenMu    sync.Mutex
	token      string
	tokenUntil time.Time
}

func NewBtnSnapClient(cfg *config.Config) *BtnSnapClient {
	return &BtnSnapClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *BtnSnapClient) GenerateQR(ctx context.Context, payload dto.BtnSnapGenerateRequest) (*dto.BtnSnapGenerateResponse, error) {
	var result dto.BtnSnapGenerateResponse
	if err := c.post(ctx, btnGeneratePath, payload, &result); err != nil {
		return nil, err
	}
	if result.ResponseCode != "2004700" {
		return nil, fmt.Errorf("BTN Generate QR rejected: %s %s", result.ResponseCode, result.ResponseMessage)
	}
	return &result, nil
}

func (c *BtnSnapClient) QueryPayment(ctx context.Context, payload dto.BtnSnapQueryRequest) (*dto.BtnSnapQueryResponse, error) {
	var result dto.BtnSnapQueryResponse
	if err := c.post(ctx, btnQueryPath, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *BtnSnapClient) post(ctx context.Context, path string, payload any, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode BTN request: %w", err)
	}

	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	now := time.Now().In(time.FixedZone("WIB", 7*60*60))
	timestamp := now.Format("2006-01-02T15:04:05-07:00")
	externalID, err := newBTNExternalID()
	if err != nil {
		return fmt.Errorf("create BTN request ID: %w", err)
	}

	hash := sha256.Sum256(body)
	stringToSign := "POST:" + path + ":" + token + ":" + hex.EncodeToString(hash[:]) + ":" + timestamp
	signature := hmac.New(sha512.New, []byte(c.cfg.BtnSnapClientSecret))
	_, _ = signature.Write([]byte(stringToSign))

	url := strings.TrimRight(c.cfg.BtnSnapBaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create BTN request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-SIGNATURE", base64.StdEncoding.EncodeToString(signature.Sum(nil)))
	req.Header.Set("X-PARTNER-ID", c.cfg.BtnSnapPartnerID)
	req.Header.Set("X-EXTERNAL-ID", externalID)
	req.Header.Set("CHANNEL-ID", c.cfg.BtnSnapChannelID)
	req.Header.Set("Origin", c.cfg.BtnSnapOrigin)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("BTN request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read BTN response: %w", err)
	}
	if err := json.Unmarshal(responseBody, result); err != nil {
		return fmt.Errorf("decode BTN response (HTTP %d): %w", resp.StatusCode, err)
	}
	return nil
}

func (c *BtnSnapClient) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.token != "" && time.Until(c.tokenUntil) > 30*time.Second {
		return c.token, nil
	}

	privateKey, err := c.privateKey()
	if err != nil {
		return "", err
	}
	now := time.Now().In(time.FixedZone("WIB", 7*60*60))
	timestamp := now.Format("2006-01-02T15:04:05-07:00")
	toSign := c.cfg.BtnSnapClientKey + "|" + timestamp
	hash := sha256.Sum256([]byte(toSign))
	signed, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign BTN token request: %w", err)
	}

	body, err := json.Marshal(dto.BtnSnapTokenRequest{
		GrantType:      "client_credentials",
		AdditionalInfo: map[string]any{},
	})
	if err != nil {
		return "", fmt.Errorf("encode BTN token request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.cfg.BtnSnapBaseURL, "/")+btnTokenPath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create BTN token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TIMESTAMP", timestamp)
	req.Header.Set("X-CLIENT-KEY", c.cfg.BtnSnapClientKey)
	req.Header.Set("X-SIGNATURE", base64.StdEncoding.EncodeToString(signed))
	req.Header.Set("Origin", c.cfg.BtnSnapOrigin)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("BTN token request failed: %w", err)
	}
	defer resp.Body.Close()

	var tokenResponse dto.BtnSnapTokenResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("decode BTN token response (HTTP %d): %w", resp.StatusCode, err)
	}
	if tokenResponse.ResponseCode != "2007300" || tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("BTN token request rejected: %s %s", tokenResponse.ResponseCode, tokenResponse.ResponseMessage)
	}
	seconds, err := strconv.Atoi(tokenResponse.ExpiresIn)
	if err != nil || seconds <= 0 {
		return "", fmt.Errorf("BTN returned invalid token expiry")
	}
	c.token = tokenResponse.AccessToken
	c.tokenUntil = now.Add(time.Duration(seconds) * time.Second)
	return c.token, nil
}

func (c *BtnSnapClient) privateKey() (*rsa.PrivateKey, error) {
	pemBytes, err := os.ReadFile(c.cfg.BtnSnapPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read BTN private key: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("decode BTN private key PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse BTN private key: %w", err)
	}
	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("BTN private key is not RSA")
	}
	return key, nil
}

func newBTNExternalID() (string, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}
