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

	"github.com/sirupsen/logrus"
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
func (dl *Dashlane) GetEntry(vault string) {

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
	XMLName            xml.Name    `xml:"root"`
	List               VaultList   `xml:"KWDataList`
	Notes              []VaultItem `xml:"KWSecureNote,omitempty"`
	Passwords          []VaultItem `xml:"KWAuthentifiant,omitempty"`
	GeneratedPasswords []VaultItem `xml:"KWGeneratedPassword,omitempty"`
	DataChangeHistory  []VaultItem `xml:"KWDataChangeHistory,omitempty"`
	Identity           []VaultItem `xml:"KWIdentity,omitempty"`
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

func (item *VaultItem) getAnonId() string {
	for _, data := range item.Datas {
		if data.Key == "AnonId" {
			return data.Value
		}
	}
	return ""
}

func (vault *Vault) Lookup(pattern string) {
	for _, item := range vault.List.Notes {
		for _, data := range item.Datas {
			if strings.Contains(data.Value, pattern) {
				fmt.Println(item.Datas)
				fmt.Println("==============")
			}
		}
	}

	for _, item := range vault.List.Passwords {
		for _, data := range item.Datas {
			if strings.Contains(data.Value, pattern) {
				fmt.Println(item.Datas)
				fmt.Println("==============")
			}
		}
	}
	for _, item := range vault.Notes {
		for _, data := range item.Datas {
			if strings.Contains(data.Value, pattern) {
				fmt.Println(item.Datas)
				fmt.Println("==============")
			}
		}
	}

	for _, item := range vault.Passwords {
		for _, data := range item.Datas {
			if strings.Contains(data.Value, pattern) {
				fmt.Println(item.Datas)
				fmt.Println("==============")
			}
		}
	}
}

func (dl *Dashlane) OpenVault(data []byte, password []byte) (*Vault, error) {
	rawVault := new(RawVault)
	err := json.Unmarshal(data, rawVault)
	if err != nil {
		return nil, err
	}

	vault := new(Vault)
	// Uncrypt the all bacup
	logrus.Debug("Opening full backup file")
	decrypted, err := dl.DecryptVault(rawVault.FullBackupFile, password)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(decrypted, vault)
	if err != nil {
		return nil, err
	}

	// Uncrypt the transactions
	logrus.Debug("Opening transactions")
	for _, transaction := range rawVault.Transactions {
		if len(transaction.Content) > 0 {
			content, err := dl.DecryptVault(transaction.Content, password)
			if err == nil {
				err := xml.Unmarshal(content, vault.List)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return vault, nil
}

/* DecryptVault decrypts the given vault with the given password
 */
func (dl *Dashlane) DecryptVault(data string, password []byte) ([]byte, error) {
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
