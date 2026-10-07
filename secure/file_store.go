package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	fileStorePort       = ":8087"
	fileStoreMaxUpload  = 16 << 20
	fileStoreMaxRequest = fileStoreMaxUpload + (256 << 10)
	fileStoreURLTTL     = 10 * time.Minute
)

var fileStoreKeyPattern = regexp.MustCompile(`^[0-9]{11}-[0-9]{5}$`)

// same signed save store as the old python server so old saves still work
func validateFileStoreConfig() error {
	if _, err := fileStoreSecret(); err != nil {
		return err
	}
	if _, err := fileStorePublicBase(); err != nil {
		return err
	}
	root := os.Getenv("FILE_SERVER_STORAGE_PATH")
	if root == "" {
		return errors.New("FILE_SERVER_STORAGE_PATH is required")
	}
	return os.MkdirAll(root, 0750)
}

func fileStoreSecret() ([]byte, error) {
	value := os.Getenv("FILE_SERVER_HMAC_SECRET")
	secret, err := hex.DecodeString(value)
	if err != nil || len(secret) != sha256.Size {
		return nil, errors.New("FILE_SERVER_HMAC_SECRET must be a 64-character hexadecimal secret")
	}
	return secret, nil
}

func fileStorePublicBase() (string, error) {
	value := strings.TrimRight(os.Getenv("FILE_SERVER_PUBLIC_URL"), "/")
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("FILE_SERVER_PUBLIC_URL must be an HTTPS origin, such as https://data.badgecade.com")
	}
	return value, nil
}

func fileStoreSignature(method, key, expires string) (string, error) {
	if !fileStoreKeyPattern.MatchString(key) || !numericExpiry(expires) {
		return "", errors.New("invalid save object key or expiry")
	}
	secret, err := fileStoreSecret()
	if err != nil {
		return "", err
	}
	message := fmt.Sprintf("badgecade-files-v1\n%s\n%s\n%s", method, key, expires)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func numericExpiry(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func fileStoreAuthorized(method, key, expires, signature string) bool {
	if !numericExpiry(expires) || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(signature) {
		return false
	}
	expiry, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		return false
	}
	now := time.Now().Unix()
	if expiry < now || expiry > now+3600 {
		return false
	}
	expected, err := fileStoreSignature(method, key, expires)
	return err == nil && hmac.Equal([]byte(expected), []byte(signature))
}

func fileStoreObjectPath(key string) (string, error) {
	if !fileStoreKeyPattern.MatchString(key) {
		return "", errors.New("invalid NEX save object key")
	}
	root, err := filepath.Abs(os.Getenv("FILE_SERVER_STORAGE_PATH"))
	if err != nil {
		return "", err
	}
	return filepath.Join(root, key), nil
}

func fileStoreObjectURL(key string) (string, uint64, error) {
	if !fileStoreKeyPattern.MatchString(key) {
		return "", 0, errors.New("invalid NEX save object key")
	}
	base, err := fileStorePublicBase()
	if err != nil {
		return "", 0, err
	}
	expires := strconv.FormatInt(time.Now().Add(fileStoreURLTTL).Unix(), 10)
	signature, err := fileStoreSignature("GET", key, expires)
	if err != nil {
		return "", 0, err
	}
	path, err := fileStoreObjectPath(key)
	if err != nil {
		return "", 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	query := url.Values{"expires": []string{expires}, "signature": []string{signature}}
	return fmt.Sprintf("%s/%s?%s", base, key, query.Encode()), uint64(info.Size()), nil
}

func fileStoreUploadInfo(key string) (string, string, string, error) {
	if !fileStoreKeyPattern.MatchString(key) {
		return "", "", "", errors.New("invalid NEX save object key")
	}
	base, err := fileStorePublicBase()
	if err != nil {
		return "", "", "", err
	}
	expires := strconv.FormatInt(time.Now().Add(fileStoreURLTTL).Unix(), 10)
	signature, err := fileStoreSignature("POST", key, expires)
	if err != nil {
		return "", "", "", err
	}
	return base + "/", expires, signature, nil
}

func startFileStore() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		body, _ := json.Marshal(map[string]string{"status": "ok"})
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write(body)
		}
	})
	mux.HandleFunc("/", handleFileStoreObject)
	logger.Success("Badgecade signed save store listening on " + fileStorePort)
	if err := http.ListenAndServe(fileStorePort, mux); err != nil {
		logger.Critical("Badgecade save store stopped: " + err.Error())
		os.Exit(1)
	}
}

func handleFileStoreObject(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		serveFileStoreDownload(w, r)
	case http.MethodPost:
		serveFileStoreUpload(w, r)
	default:
		w.Header().Set("Allow", "GET, HEAD, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func serveFileStoreDownload(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/")
	query := r.URL.Query()
	if strings.Contains(key, "/") || len(query) != 2 || len(query["expires"]) != 1 || len(query["signature"]) != 1 || !fileStoreAuthorized("GET", key, query.Get("expires"), query.Get("signature")) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	path, err := fileStoreObjectPath(key)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = io.Copy(w, file)
	}
}

func serveFileStoreUpload(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.URL.RawQuery != "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, fileStoreMaxRequest)
	values, fileData, err := readFileStoreUpload(r)
	if err != nil {
		logger.Error(fmt.Sprintf("Rejected save upload (Content-Type %q, Content-Length %d): %s", r.Header.Get("Content-Type"), r.ContentLength, err))
		http.Error(w, "invalid upload", http.StatusRequestEntityTooLarge)
		return
	}
	keyValues := values["key"]
	expiresValues := values["expires"]
	signatureValues := values["signature"]
	aclValues := values["acl"]
	if len(keyValues) != 1 || len(expiresValues) != 1 || len(signatureValues) != 1 || len(aclValues) != 1 || aclValues[0] != "private" {
		logger.Error(fmt.Sprintf("Rejected save upload: missing or repeated form fields (have %v)", fileStoreFieldNames(values)))
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	key, expires, signature := keyValues[0], expiresValues[0], signatureValues[0]
	if !fileStoreAuthorized("POST", key, expires, signature) {
		logger.Error("Rejected save upload: invalid or expired signature for " + key)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	source := io.NopCloser(bytes.NewReader(fileData))
	defer source.Close()
	destination, err := fileStoreObjectPath(key)
	if err != nil {
		http.Error(w, "invalid object key", http.StatusBadRequest)
		return
	}
	tmp, err := os.CreateTemp(filepath.Dir(destination), ".badgecade-upload-*")
	if err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0644); err != nil {
		_ = tmp.Close()
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	written, copyErr := io.Copy(tmp, io.LimitReader(source, fileStoreMaxUpload+1))
	closeErr := tmp.Close()
	if copyErr != nil || closeErr != nil || written > fileStoreMaxUpload {
		http.Error(w, "invalid upload", http.StatusRequestEntityTooLarge)
		return
	}
	if err := os.Rename(tmpName, destination); err != nil {
		http.Error(w, "storage unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNoContent)
}

// read the multipart parts by hand, the 3DS sometimes leaves out the filename
// and ParseMultipartForm doesn't treat it as a file then
func readFileStoreUpload(r *http.Request) (map[string][]string, []byte, error) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, nil, fmt.Errorf("bad Content-Type: %w", err)
	}
	if !strings.HasPrefix(mediaType, "multipart/") || params["boundary"] == "" {
		return nil, nil, fmt.Errorf("not a multipart upload (%s)", mediaType)
	}

	values := map[string][]string{}
	var fileData []byte
	haveFile := false
	reader := multipart.NewReader(r.Body, params["boundary"])
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("reading multipart: %w", err)
		}
		data, err := io.ReadAll(io.LimitReader(part, fileStoreMaxUpload+1))
		name := part.FormName()
		isFile := name == "file" || part.FileName() != ""
		_ = part.Close()
		if err != nil {
			return nil, nil, fmt.Errorf("reading part %q: %w", name, err)
		}
		if isFile {
			if haveFile {
				return nil, nil, errors.New("more than one file part")
			}
			if len(data) > fileStoreMaxUpload {
				return nil, nil, errors.New("save is larger than the upload limit")
			}
			fileData, haveFile = data, true
			continue
		}
		if len(data) > 4096 {
			return nil, nil, fmt.Errorf("form field %q is too large", name)
		}
		values[name] = append(values[name], string(data))
	}
	if !haveFile {
		return nil, nil, fmt.Errorf("no file part (fields: %v)", fileStoreFieldNames(values))
	}
	return values, fileData, nil
}

func fileStoreFieldNames(values map[string][]string) []string {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	return names
}
