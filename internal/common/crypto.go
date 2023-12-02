package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/sweet-go/stdlib/encryption"
	"golang.org/x/crypto/bcrypt"
)

// DefaultBlockSize is the default block size used for encryption/decryption
const DefaultBlockSize int = 16

// SharedCryptor hold useful function to perform encryption related prosess, such as encrypting email, hashing password, etc
type SharedCryptor interface {
	Encrypt(plainText string) (encryptedText string, err error)
	Decrypt(cipherText string) (plainText string, err error)
	Hash(data []byte) (string, error)
	CompareHash(hashed []byte, plain []byte) error
	CreateSecureToken() (string, string, error)
	ReverseSecureToken(plain string) string
}

// CreateCryptorOpts is the options used to create a new cryptor instance.
type CreateCryptorOpts struct {
	HashCost      int
	EncryptionKey []byte
	IV            string
	BlockSize     int
}

type sharedCryptor struct {
	encryptionKey []byte
	iv            string
	blockSize     int
}

// NewSharedCryptor returns a new instance of SharedCryptor
func NewSharedCryptor(opts *CreateCryptorOpts) SharedCryptor {
	return &sharedCryptor{
		encryptionKey: encryption.SHA256Hash(opts.EncryptionKey),
		iv:            opts.IV,
		blockSize:     opts.BlockSize,
	}
}

func (s *sharedCryptor) Encrypt(plainText string) (string, error) {
	ivKey, err := hex.DecodeString(s.iv)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	bPlaintext := s.pkcs5Padding([]byte(plainText), s.blockSize, len(plainText))
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		logrus.Error(err)
		return "", err
	}

	ciphertext := make([]byte, len(bPlaintext))

	mode := cipher.NewCBCEncrypter(block, ivKey)
	mode.CryptBlocks(ciphertext, bPlaintext)

	return hex.EncodeToString(ciphertext), nil
}

func (s *sharedCryptor) Decrypt(cipherText string) (plainText string, err error) {
	ivKey, err := s.generateIVKey(s.iv)
	if err != nil {
		return
	}

	cipherTextDecoded, err := hex.DecodeString(cipherText)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return
	}

	mode := cipher.NewCBCDecrypter(block, ivKey)
	mode.CryptBlocks(cipherTextDecoded, cipherTextDecoded)

	return string(s.pkcs5Unpadding(cipherTextDecoded)), nil
}

func (s *sharedCryptor) Hash(data []byte) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hashed), nil
}

func (s *sharedCryptor) CompareHash(hashed []byte, plain []byte) error {
	return bcrypt.CompareHashAndPassword(hashed, plain)
}

func (s *sharedCryptor) CreateSecureToken() (string, string, error) {
	random := []byte(uuid.New().String())
	raw := uuid.NewHash(sha256.New(), uuid.New(), random, 5)

	plain, err := s.Hash([]byte(raw.String()))
	if err != nil {
		return "", "", err
	}

	enc := encryption.SHA256Hash([]byte(plain))
	encoded := base64.StdEncoding.EncodeToString(enc)

	return plain, encoded, nil
}

func (s *sharedCryptor) ReverseSecureToken(plain string) string {
	tokenEnc := encryption.SHA256Hash([]byte(plain))
	encoded := base64.StdEncoding.EncodeToString(tokenEnc)

	return encoded
}

func (s *sharedCryptor) generateIVKey(iv string) (bIv []byte, err error) {
	if len(iv) > 0 {
		ivKey, err := hex.DecodeString(iv)
		if err != nil {
			return nil, fmt.Errorf("unable to hex decode iv")
		}
		return ivKey, nil
	}

	ivKey, err := generateRandomIVKey(s.blockSize)
	if err != nil {
		return nil, fmt.Errorf("unable to generate random iv key")
	}

	return hex.DecodeString(ivKey)
}

func (s *sharedCryptor) pkcs5Padding(ciphertext []byte, blockSize int, _ int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	return append(ciphertext, padtext...)
}

func (s *sharedCryptor) pkcs5Unpadding(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	length := len(src)
	unpadding := int(src[length-1])
	cutLen := (length - unpadding)
	// check boundaries
	if cutLen < 0 || cutLen > length {
		return src
	}

	return src[:cutLen]
}

// GenerateRandomIVKey generate random IV value
func generateRandomIVKey(blockSize int) (bIv string, err error) {
	bytes := make([]byte, blockSize)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
