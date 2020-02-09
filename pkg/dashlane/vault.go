package dashlane

import (
	"bytes"
	"compress/flate"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
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

type VaultEntry struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

type VaultItem struct {
	Datas []VaultEntry `xml:"KWDataItem"`
}

type VaultList struct {
	XMLName   xml.Name    `xml:"KWDataList"`
	Passwords []VaultItem `xml:"KWAuthentifiant,omitempty"`
	Notes     []VaultItem `xml:"KWSecureNote,omitempty"`
}

type Vault struct {
	XMLName xml.Name  `xml:"root"`
	List    VaultList `xml:"KWDataList`
}

type TransactionsEntry struct {
	Action     string `json:"action"`
	BackupDate int    `json:"backupdate"`
	Content    string `json:"content,omitempty"`
	Identifier string `json:"identifier"`
	ObjectType string `json:"objectType"`
	Time       int    `json:"time"`
	Type       string `json:"type"`
}

type RawVault struct {
	Transactions   []TransactionsEntry `json:"transactionList"`
	FullBackupFile string              `json:"fullBackupFile"`
}

func LoadVault(data []byte) (*RawVault, error) {
	rawVault := new(RawVault)
	err := json.Unmarshal(data, rawVault)
	if err != nil {
		return nil, err
	}
	return rawVault, nil
}

func ParseVault(data string, password []byte) (*Vault, error) {
	decrypted, err := DecryptVault(data, password)
	if err != nil {
		return nil, err
	}
	// fmt.Println(string(decrypted))
	vault := new(Vault)
	err = xml.Unmarshal(decrypted, vault)
	if err != nil {
		return nil, err
	}

	return vault, nil
}

/* DecryptVault decrypts the given vault with the given password
 */
func DecryptVault(data string, password []byte) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	encryptedData := parseEncryptedData(string(decoded))

	originalKey := deriveEncryptionKey(encryptedData, password)
	ivKey, iv := deriveEncryptionIV(encryptedData, originalKey)
	key := originalKey
	if encryptedData.useDerivedKey {
		key = ivKey
	}
	plaintext, err := uncrypt(encryptedData.ciphertext, iv, key)
	if err != nil {
		return nil, err
	}

	if encryptedData.compressed {
		return uncompress(plaintext[6:len(plaintext)])
	}
	return plaintext, nil
}

func uncompress(data []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func uncrypt(ciphertext string, iv, key []byte) ([]byte, error) {
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext (%v) is not a multiple of the block size (%v)", len(ciphertext), aes.BlockSize)
	}
	if len(key)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("key size not multiple of block size")
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("iv size not same as block size")
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

func deriveEncryptionIV(data *encryptedData, key []byte) ([]byte, []byte) {
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

func deriveEncryptionKey(data *encryptedData, password []byte) []byte {
	return pbkdf2.Key(password, []byte(data.salt), 10204, 32, sha1.New)
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
