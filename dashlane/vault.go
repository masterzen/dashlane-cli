package dashlane

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLength    = 32
	versionLength = 4
	kwc3Version   = "KWC3"
)

type encryptedData struct {
	salt          string
	ciphertext    string
	compressed    bool
	useDerivedKey bool
	iterations    int
	version       string
}

/*
 */
func GetEntry(vault string) {

}

type PasswordEntry struct {
	ID       string `xml:"Id,attr"`
	Title    string `xml:"Title,attr"`
	Login    string `xml:"Login,attr"`
	Password string `xml:"Password,attr"`
}

type Vault struct {
	Passwords []PasswordEntry `xml:"KWAuthentifiant>KWDataItem"`
}

func ParseVault(data, password string) (Vault, error) {
	decrypted, err := DecryptVault(data, password)
	if err != nil {
		return err
	}

	// parse the XML

}

/* DecryptVault decrypts the given vault with the given password
 */
func DecryptVault(data string, password string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	encryptedData := parseEncryptedData(string(decoded))

	originalKey := deriveEncryptionKey(encryptedData, password)
	ivKey, iv := deriveEncryptionIV(encryptedData, originalKey, password)
	key := originalKey
	if encryptedData.useDerivedKey {
		key = ivKey
	}
	plaintext, err := uncrypt(encryptedData.ciphertext, iv, key)
	if err != nil {
		return "", err
	}

	if encryptedData.compressed {
		return uncompress(plaintext[6:len(plaintext)])
	}
	return string(plaintext), nil
}

func uncompress(data []byte) (string, error) {
	fmt.Println("plaintext")
	fmt.Print(hex.Dump(data))
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err == nil {
		return string(b), nil
	}
	return "", err
}

func uncrypt(ciphertext string, iv, key []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		panic(fmt.Sprintf("ciphertext (%v) is not a multiple of the block size (%v)", len(ciphertext), aes.BlockSize))
	}
	if len(key)%aes.BlockSize != 0 {
		panic("key size not multiple of block size")
	}
	if len(iv) != aes.BlockSize {
		panic("iv size not same as block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, []byte(ciphertext))
	return plaintext, nil
}

func deriveEncryptionIV(data *encryptedData, key []byte, password string) ([]byte, []byte) {
	salted := append(key, []byte(data.salt[:8])...)

	parts := []string{""}
	for i := 0; i < 3; i++ {
		appended := append([]byte(parts[len(parts)-1]), salted...)
		sha1ed := multipleSha1(appended, data.iterations)
		parts = append(parts, string(sha1ed))
	}
	keyIV := strings.Join(parts, "")

	return []byte(keyIV[0:32]), []byte(keyIV[32:48])
}

func deriveEncryptionKey(data *encryptedData, password string) []byte {
	return pbkdf2.Key([]byte(password), []byte(data.salt), 10204, 32, sha1.New)
}

func multipleSha1(b []byte, it int) []byte {
	for j := 0; j < it; j++ {
		h := sha1.Sum(b)
		b = h[:]
	}
	return b
}

func parseEncryptedData(data string) *encryptedData {
	salt := data[0:saltLength]
	version := data[saltLength : saltLength+versionLength]

	if version == kwc3Version {
		return &encryptedData{
			salt:          salt,
			ciphertext:    data[saltLength+versionLength : len(data)],
			compressed:    true,
			useDerivedKey: false,
			iterations:    1,
			version:       version,
		}
	}

	return &encryptedData{
		salt:          salt,
		ciphertext:    data[saltLength:len(data)],
		compressed:    false,
		useDerivedKey: true,
		iterations:    5,
		version:       version,
	}
}
