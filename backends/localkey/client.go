package localkey

import (
	"errors"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/secrets-api/pkg/aesutils"
)

// Client implements the backend client interface
type Client struct {
	encryptionKeyPath string
}

// support both IV and Nonce for non-breaking
type internalSecret struct {
	Nonce      []byte `json:"nonce,omitempty"`
	IV         []byte `json:"iv,ommitempty"`
	Algorithm  string
	CipherText []byte
}

// NewLocalKey initializes a new local key
func NewLocalKey(keyPath string) (*Client, error) {
	err := errors.New("No encryption key path configured. Must be a directory")

	if keyPath != "" {
		if isDir, err := testIsDir(keyPath); isDir && err == nil {
			return &Client{encryptionKeyPath: keyPath}, nil
		}
	}

	return &Client{}, err
}

func (l *Client) loadEncryptionKeyFromPath(keyName string) (aesutils.AESKey, error) {
	keyFile := path.Join(l.encryptionKeyPath, keyName)

	return aesutils.NewAESKeyFromFile(keyFile)
}

// GetEncryptedText localkey Client just returns the clearText
func (l *Client) GetEncryptedText(keyName, clearText string) (string, error) {
	key, err := l.loadEncryptionKeyFromPath(keyName)
	if err != nil {
		return "", err
	}

	return aesutils.GetEncryptedText(key, clearText, "aes256-gcm")
}

// GetClearText localkey Client
func (l *Client) GetClearText(keyName, secretBlob string) (string, error) {
	key, err := l.loadEncryptionKeyFromPath(keyName)
	if err != nil {
		return "", err
	}

	return aesutils.GetClearText(key, secretBlob)
}

// Sign implements the interface
func (l *Client) Sign(keyName, clearText string) (string, error) {
	key, err := l.loadEncryptionKeyFromPath(keyName)
	if err != nil {
		return "", err
	}

	return aesutils.Sign(key, clearText)
}

// VerifySignature implements the interface.
func (l *Client) VerifySignature(keyName, signature, message string) (bool, error) {
	key, err := l.loadEncryptionKeyFromPath(keyName)
	if err != nil {
		return false, err
	}

	return aesutils.VerifySignature(key, signature, message)
}

// Delete No op nothing stored
func (l *Client) Delete(keyName, cipherText string) error {
	return nil
}

func testIsDir(keyPath string) (bool, error) {
	result := false

	file, err := os.Open(keyPath)
	if err != nil {
		logrus.Error(err)
		return result, err
	}
	defer file.Close()

	fs, err := file.Stat()
	if err != nil {
		return result, err
	}

	return fs.IsDir(), nil
}
