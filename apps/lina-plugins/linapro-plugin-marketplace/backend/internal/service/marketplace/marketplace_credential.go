// This file stores and decrypts private Git access tokens for marketplace
// metadata discovery. Ciphertext is never returned through marketplace APIs.

package marketplace

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"strings"

	"lina-core/pkg/bizerr"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const marketplaceCredentialKeyEnv = "LINAPRO_MARKETPLACE_CREDENTIAL_KEY"

// saveGitCredential encrypts one access token and returns an opaque credential ref.
func (s *serviceImpl) saveGitCredential(
	ctx context.Context,
	ownerUserID int64,
	provider string,
	accessToken string,
) (string, error) {
	token := strings.TrimSpace(accessToken)
	if token == "" {
		return "", nil
	}
	if ownerUserID <= 0 {
		return "", bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	cipherText, err := encryptMarketplaceSecret(token)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	refBytes := make([]byte, 16)
	if _, err = rand.Read(refBytes); err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	ref := "mpcred_" + hex.EncodeToString(refBytes)
	if _, err = dao.PluginMarketplaceCredential.Ctx(ctx).Data(do.PluginMarketplaceCredential{
		CredentialRef: ref,
		OwnerUserId:   ownerUserID,
		Provider:      normalizeKey(provider),
		CipherText:    cipherText,
	}).Insert(); err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return ref, nil
}

// loadGitCredentialToken decrypts one credential for metadata discovery only.
func (s *serviceImpl) loadGitCredentialToken(ctx context.Context, credentialRef string) (string, error) {
	ref := normalizeKey(credentialRef)
	if ref == "" {
		return "", nil
	}
	var row *entity.PluginMarketplaceCredential
	if err := dao.PluginMarketplaceCredential.Ctx(ctx).
		Where(do.PluginMarketplaceCredential{CredentialRef: ref}).
		Scan(&row); err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if row == nil || strings.TrimSpace(row.CipherText) == "" {
		return "", nil
	}
	token, err := decryptMarketplaceSecret(row.CipherText)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return token, nil
}

// encryptMarketplaceSecret seals plaintext with AES-GCM using a platform key.
func encryptMarketplaceSecret(plaintext string) (string, error) {
	block, err := aes.NewCipher(marketplaceCredentialKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// decryptMarketplaceSecret opens ciphertext sealed by encryptMarketplaceSecret.
func decryptMarketplaceSecret(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(marketplaceCredentialKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", bizerr.NewCode(CodeMarketplaceStorageFailed)
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// marketplaceCredentialKey derives a 32-byte AES key from env or a dev default.
func marketplaceCredentialKey() []byte {
	seed := strings.TrimSpace(os.Getenv(marketplaceCredentialKeyEnv))
	if seed == "" {
		// Development fallback keeps local installs working; production should set the env key.
		seed = "linapro-marketplace-dev-credential-key"
	}
	sum := sha256.Sum256([]byte(seed))
	return sum[:]
}
