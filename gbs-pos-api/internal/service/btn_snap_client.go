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
	btnTokenPath      = "/snap/v1/access-token/b2b"
	btnGeneratePath   = "/snap/v1/qr/qr-mpm-generate"
	btnQueryPath      = "/snap/v1/qr/qr-mpm-query"
	btnCPMDecodePath  = "/snap/v1.0/qr/qr-cpm-decode"
	btnCPMPaymentPath = "/snap/v1.0/qr/qr-cpm-payment"
	btnCPMQueryPath   = "/snap/v1.0/qr/qr-cpm-query"
	btnCPMCancelPath  = "/snap/v1.0/qr/qr-cpm-cancel"
	btnCPMRefundPath  = "/snap/v1.0/qr/qr-cpm-refund"
	btnMPMRefundPath  = "/snap/v1.0/qr/qr-mpm-refund"
	btnMPMReportPath  = "/snap/v1.0/qr/qr-mpm-report"
)

type BtnSnapClient struct {
	cfg        *config.Config
	httpClient *http.Client
	tokenMu    sync.Mutex
	token      string
	tokenUntil time.Time
}

// BtnSnapRejectedError is a definitive Generate response that did not create a QR payment.
type BtnSnapRejectedError struct {
	ResponseCode    string
	ResponseMessage string
}

func (e *BtnSnapRejectedError) Error() string {
	return fmt.Sprintf("BTN Generate QR rejected: %s %s", e.ResponseCode, e.ResponseMessage)
}

// BtnSnapOutcomeUnknownError means a payment request may have reached BTN but its result could not be verified.
type BtnSnapOutcomeUnknownError struct {
	Cause error
}

type BtnSnapCPMRejectedError struct {
	Operation       string
	ResponseCode    string
	ResponseMessage string
}

func (e *BtnSnapCPMRejectedError) Error() string {
	return fmt.Sprintf("BTN CPM %s rejected: %s %s", e.Operation, e.ResponseCode, e.ResponseMessage)
}

func (e *BtnSnapOutcomeUnknownError) Error() string {
	return fmt.Sprintf("BTN payment outcome is unknown; reconcile using the same reference: %v", e.Cause)
}

func (e *BtnSnapOutcomeUnknownError) Unwrap() error { return e.Cause }

func NewBtnSnapClient(cfg *config.Config) *BtnSnapClient {
	return &BtnSnapClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *BtnSnapClient) GenerateQR(ctx context.Context, payload dto.BtnSnapGenerateRequest) (*dto.BtnSnapGenerateResponse, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	return c.generateQRWithToken(ctx, payload, token)
}

func (c *BtnSnapClient) generateQRWithToken(ctx context.Context, payload dto.BtnSnapGenerateRequest, token string) (*dto.BtnSnapGenerateResponse, error) {
	var result dto.BtnSnapGenerateResponse
	statusCode, err := c.postWithToken(ctx, btnGeneratePath, payload, &result, token)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusOK && result.ResponseCode == "2004700" {
		return &result, nil
	}
	if definitiveBTNGenerateRejection(statusCode, result.ResponseCode) {
		return nil, &BtnSnapRejectedError{ResponseCode: result.ResponseCode, ResponseMessage: result.ResponseMessage}
	}
	return nil, fmt.Errorf("BTN Generate response could not be verified (HTTP %d, responseCode %q)", statusCode, result.ResponseCode)
}

func (c *BtnSnapClient) QueryPayment(ctx context.Context, payload dto.BtnSnapQueryRequest) (*dto.BtnSnapQueryResponse, error) {
	var result dto.BtnSnapQueryResponse
	statusCode, err := c.post(ctx, btnQueryPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound && result.ResponseCode == "4045101" {
		return &result, nil
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("BTN Query returned HTTP %d (responseCode %q)", statusCode, result.ResponseCode)
	}
	return &result, nil
}

func (c *BtnSnapClient) DecodeCPM(ctx context.Context, payload dto.BtnSnapCPMDecodeRequest) (*dto.BtnSnapCPMDecodeResponse, error) {
	var result dto.BtnSnapCPMDecodeResponse
	status, err := c.post(ctx, btnCPMDecodePath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2000000" {
		return nil, &BtnSnapCPMRejectedError{Operation: "Decode", ResponseCode: result.ResponseCode, ResponseMessage: result.ResponseMessage}
	}
	return &result, nil
}

func (c *BtnSnapClient) PayCPM(ctx context.Context, payload dto.BtnSnapCPMPaymentRequest) (*dto.BtnSnapCPMPaymentResponse, error) {
	var result dto.BtnSnapCPMPaymentResponse
	status, err := c.post(ctx, btnCPMPaymentPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2006000" {
		if status >= 400 && status < 500 && status != http.StatusConflict && result.ResponseCode != "" {
			return nil, &BtnSnapCPMRejectedError{Operation: "Payment", ResponseCode: result.ResponseCode, ResponseMessage: result.ResponseMessage}
		}
		return nil, fmt.Errorf("BTN CPM Payment result could not be verified (HTTP %d, responseCode %q)", status, result.ResponseCode)
	}
	return &result, nil
}

func (c *BtnSnapClient) QueryCPM(ctx context.Context, payload dto.BtnSnapCPMQueryRequest) (*dto.BtnSnapCPMQueryResponse, error) {
	var result dto.BtnSnapCPMQueryResponse
	status, err := c.post(ctx, btnCPMQueryPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2006100" {
		return nil, fmt.Errorf("BTN CPM Query rejected: HTTP %d, %s %s", status, result.ResponseCode, result.ResponseMessage)
	}
	return &result, nil
}

func (c *BtnSnapClient) CancelCPM(ctx context.Context, payload dto.BtnSnapCPMCancelRequest) (*dto.BtnSnapCPMCancelResponse, error) {
	var result dto.BtnSnapCPMCancelResponse
	status, err := c.post(ctx, btnCPMCancelPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2006200" {
		return nil, fmt.Errorf("BTN CPM Cancel result could not be verified (HTTP %d, responseCode %q)", status, result.ResponseCode)
	}
	return &result, nil
}

func (c *BtnSnapClient) RefundCPM(ctx context.Context, payload dto.BtnSnapCPMRefundRequest) (*dto.BtnSnapCPMRefundResponse, error) {
	var result dto.BtnSnapCPMRefundResponse
	status, err := c.post(ctx, btnCPMRefundPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2008000" {
		return nil, fmt.Errorf("BTN CPM Refund result could not be verified (HTTP %d, responseCode %q)", status, result.ResponseCode)
	}
	return &result, nil
}

func (c *BtnSnapClient) RefundMPM(ctx context.Context, payload dto.BtnSnapMPMRefundRequest) (*dto.BtnSnapMPMRefundResponse, error) {
	var result dto.BtnSnapMPMRefundResponse
	status, err := c.post(ctx, btnMPMRefundPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2007800" {
		return nil, fmt.Errorf("BTN MPM Refund result could not be verified (HTTP %d, responseCode %q)", status, result.ResponseCode)
	}
	return &result, nil
}

func (c *BtnSnapClient) ReportMPM(ctx context.Context, payload dto.BtnSnapMPMReportRequest) (*dto.BtnSnapMPMReportResponse, error) {
	var result dto.BtnSnapMPMReportResponse
	status, err := c.post(ctx, btnMPMReportPath, payload, &result)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK || result.ResponseCode != "2009900" {
		return nil, fmt.Errorf("BTN MPM Report rejected: HTTP %d, %s %s", status, result.ResponseCode, result.ResponseMessage)
	}
	return &result, nil
}

func (c *BtnSnapClient) post(ctx context.Context, path string, payload any, result any) (int, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return 0, err
	}
	return c.postWithToken(ctx, path, payload, result, token)
}

func (c *BtnSnapClient) postWithToken(ctx context.Context, path string, payload any, result any, token string) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("encode BTN request: %w", err)
	}

	now := time.Now().In(time.FixedZone("WIB", 7*60*60))
	timestamp := now.Format("2006-01-02T15:04:05-07:00")
	externalID, err := newBTNExternalID()
	if err != nil {
		return 0, fmt.Errorf("create BTN request ID: %w", err)
	}

	hash := sha256.Sum256(body)
	stringToSign := "POST:" + path + ":" + token + ":" + hex.EncodeToString(hash[:]) + ":" + timestamp
	signature := hmac.New(sha512.New, []byte(c.cfg.BtnSnapClientSecret))
	_, _ = signature.Write([]byte(stringToSign))

	url := strings.TrimRight(c.cfg.BtnSnapBaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create BTN request: %w", err)
	}
	req.Header["Authorization"] = []string{"Bearer " + token}
	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["X-TIMESTAMP"] = []string{timestamp}
	req.Header["X-SIGNATURE"] = []string{base64.StdEncoding.EncodeToString(signature.Sum(nil))}
	req.Header["X-PARTNER-ID"] = []string{c.cfg.BtnSnapPartnerID}
	req.Header["X-EXTERNAL-ID"] = []string{externalID}
	req.Header["CHANNEL-ID"] = []string{c.cfg.BtnSnapChannelID}
	req.Header["Origin"] = []string{c.cfg.BtnSnapOrigin}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("BTN request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read BTN response: %w", err)
	}
	if err := json.Unmarshal(responseBody, result); err != nil {
		return resp.StatusCode, fmt.Errorf("decode BTN response (HTTP %d): %w", resp.StatusCode, err)
	}
	return resp.StatusCode, nil
}

func definitiveBTNGenerateRejection(statusCode int, responseCode string) bool {
	if len(responseCode) != 7 || responseCode[3:5] != "47" {
		return false
	}
	for _, digit := range responseCode {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	httpCode, err := strconv.Atoi(responseCode[:3])
	if err != nil || httpCode != statusCode {
		return false
	}
	switch httpCode {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusTooManyRequests:
		return true
	default:
		return false
	}
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
	req.Header["Content-Type"] = []string{"application/json"}
	req.Header["X-TIMESTAMP"] = []string{timestamp}
	req.Header["X-CLIENT-KEY"] = []string{c.cfg.BtnSnapClientKey}
	req.Header["X-SIGNATURE"] = []string{base64.StdEncoding.EncodeToString(signed)}
	req.Header["Origin"] = []string{c.cfg.BtnSnapOrigin}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("BTN token request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var rejection dto.BtnSnapTokenResponse
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rejection)
		if rejection.ResponseCode != "" {
			return "", fmt.Errorf("BTN token request rejected: HTTP %d, %s %s", resp.StatusCode, rejection.ResponseCode, rejection.ResponseMessage)
		}
		return "", fmt.Errorf("BTN token request returned HTTP %d", resp.StatusCode)
	}

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
