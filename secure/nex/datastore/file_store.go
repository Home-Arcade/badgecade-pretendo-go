package nex_datastore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/PretendoNetwork/nex-protocols-go/datastore"
)

var objectKeyPattern = regexp.MustCompile(`^[0-9]{11}-[0-9]{5}$`)

func objectKey(dataID, version uint32) string {
	return fmt.Sprintf("%011d-%05d", dataID, version)
}

func signObject(method, key, expires string) (string, error) {
	if !objectKeyPattern.MatchString(key) {
		return "", errors.New("invalid NEX save object key")
	}
	secret, err := hex.DecodeString(os.Getenv("FILE_SERVER_HMAC_SECRET"))
	if err != nil || len(secret) != sha256.Size {
		return "", errors.New("FILE_SERVER_HMAC_SECRET must be a 64-character hexadecimal secret")
	}
	message := fmt.Sprintf("badgecade-files-v1\n%s\n%s\n%s", method, key, expires)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func makeObjectDownloadURL(key string) (string, uint64, error) {
	if !objectKeyPattern.MatchString(key) {
		return "", 0, errors.New("invalid NEX save object key")
	}
	expires := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	signature, err := signObject("GET", key, expires)
	if err != nil {
		return "", 0, err
	}
	path := filepath.Join(os.Getenv("FILE_SERVER_STORAGE_PATH"), key)
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	base := os.Getenv("FILE_SERVER_PUBLIC_URL")
	query := url.Values{"expires": []string{expires}, "signature": []string{signature}}
	return fmt.Sprintf("%s/%s?%s", base, key, query.Encode()), uint64(info.Size()), nil
}

func makeObjectUploadInfo(key string) (string, []*datastore.DataStoreKeyValue, error) {
	if !objectKeyPattern.MatchString(key) {
		return "", nil, errors.New("invalid NEX save object key")
	}
	expires := strconv.FormatInt(time.Now().Add(10*time.Minute).Unix(), 10)
	signature, err := signObject("POST", key, expires)
	if err != nil {
		return "", nil, err
	}
	fields := []*datastore.DataStoreKeyValue{
		objectFormField("key", key),
		objectFormField("acl", "private"),
		objectFormField("expires", expires),
		objectFormField("signature", signature),
	}
	return os.Getenv("FILE_SERVER_PUBLIC_URL") + "/", fields, nil
}

func objectFormField(key, value string) *datastore.DataStoreKeyValue {
	field := datastore.NewDataStoreKeyValue()
	field.Key = key
	field.Value = value
	return field
}
