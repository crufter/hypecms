package user_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/basic"
	ifaces "github.com/opesun/hypecms/interfaces"
	"crypto/sha1"
	"crypto/cipher"
	"crypto/rand"
	"crypto/aes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"errors"
)

const(
	block_size = 16		// For encryption and decryption.
)

func FindUser(db *mgo.Database, id string) (map[string]interface{}, error) {
	v:= basic.Find(db, "users", id)
	if v != nil {
		delete(v, "password")
		return v, nil
	}
	return nil, fmt.Errorf("Can't find user with id " + id)
}

func NamePass(db *mgo.Database, name, encoded_pass string) (map[string]interface{}, error) {
	var v interface{}
	db.C("users").Find(bson.M{"name": name, "password": encoded_pass}).One(&v)
	if v != nil {
		return basic.Convert(v).(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("Can't find user/password combo.")
}

func Login(db *mgo.Database, inp map[string][]string, block_key []byte) (map[string]interface{}, string, error) {
	if len(block_key) < block_size {
		return nil, "", fmt.Errorf("Login: block_key length must be at least %v.", block_size)
	}
	block_key = block_key[0:block_size]
	rule := map[string]interface{}{
		"name": "must",
		"password": "must",
	}
	d, err := extract.New(rule).Extract(inp)
	if err != nil {
		return nil, "", err
	}
	user, err := NamePass(db, d["name"].(string), Encode(d["password"].(string)))
	if err != nil {
		return nil, "", err
	}
	id_b, err := encryptStr(block_key, user["_id"].(bson.ObjectId).Hex())
	if err != nil { return nil, "", err }
	return user, string(id_b), nil
}

func EmptyUser() map[string]interface{} {
	user := make(map[string]interface{})
	user["level"] = 0
	return user
}

func ParseAcceptLanguage(l string) []string {
	ret := []string{}
	sl := strings.Split(l, ",")
	c := map[string]struct{}{}
	for _, v := range sl {
		lang := string(strings.Split(v, ";")[0][0:2])
		_, has := c[lang]
		if !has {
			c[lang] = struct{}{}
			ret = append(ret, lang)
		}
	}
	return ret
}

func BuildUser(db *mgo.Database, ev ifaces.Event, user_id string, http_header map[string][]string, block_key []byte) (map[string]interface{}, error) {
	if len(block_key) < 16 {
		return nil, fmt.Errorf("BuildUser: block_key length must be at least %v.", block_size)
	}
	block_key = block_key[0:block_size]
	var user map[string]interface{}
	var err error
	if len(user_id) > 0 {
		decr_id_b, err := decryptStr(block_key, user_id)
		if err != nil { return nil, err }
		user, err = FindUser(db, string(decr_id_b))
	}
	if err != nil || user == nil {
		user = EmptyUser()
	}
	_, langs_are_set := user["languages"]
	if !langs_are_set {
		langs, has := http_header["Accept-Language"]
		if has {
			user["languages"] = ParseAcceptLanguage(langs[0])
		} else {
			user["languages"] = []string{"en"}
		}
	}
	ev.Trigger("user.build", user)
	return user, nil
}

func Encode(pass string) string {
	h := sha1.New()
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// We should call the extract module here, also no name && pass but rather a map[string]interface{} containing all the things.
func Register(db *mgo.Database, ev ifaces.Event, name, pass string) error {
	u := bson.M{"name": name, "password": Encode(pass)}
	err := db.C("users").Insert(u)
	if err != nil {
		return fmt.Errorf("Name is not unique.")
	}
	ev.Trigger("user.register", u)
	return nil
}

// Function intended to encrypt the user id before storing it as a cookie.
// encr flag controls
// block_key is the salt.
func encDecStr(block_key[]byte, value string, encr bool) (string, error) {
	if block_key == nil || len(block_key) == 0 { return "", fmt.Errorf("Can't encrypt/decrypt: block key is not proper.") }
	block, err := aes.NewCipher(block_key)
	if err != nil { return "", err }
	var bs []byte
	if encr {
		bs, err = encrypt(block, []byte(value))
	} else {
		bs, err = decrypt(block, []byte(value))
	}
	if err != nil { return "", err }
	if bs == nil { return "", fmt.Errorf("Somethign went wrong when encoding/decoding.") } // Just in case.
	return string(bs), nil
}

func encryptStr(block_key []byte, value string) (string, error) {
	str, err := encDecStr(block_key, value, true)
	if err != nil { return "", err }
	return base64.StdEncoding.EncodeToString([]byte(str)), nil
}

func decryptStr(block_key []byte, value string) (string, error) {
	decoded_b, err := base64.StdEncoding.DecodeString(value)
	if err != nil { return "", err }
	return encDecStr(block_key, string(decoded_b), false)
}

// following functions are taken from securecookie package of the Gorilla web toolkit made by Rodrigo Moraes.

// encrypt encrypts a value using the given block in counter mode.
//
// A random initialization vector (http://goo.gl/zF67k) with the length of the
// block size is prepended to the resulting ciphertext.
func encrypt(block cipher.Block, value []byte) ([]byte, error) {
	iv := GenerateRandomKey(block.BlockSize())
	if iv == nil {
		return nil, errors.New("securecookie: failed to generate random iv")
	}
	// Encrypt it.
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(value, value)
	// Return iv + ciphertext.
	return append(iv, value...), nil
}

// decrypt decrypts a value using the given block in counter mode.
//
// The value to be decrypted must be prepended by a initialization vector
// (http://goo.gl/zF67k) with the length of the block size.
func decrypt(block cipher.Block, value []byte) ([]byte, error) {
	size := block.BlockSize()
	if len(value) > size {
		// Extract iv.
		iv := value[:size]
		// Extract ciphertext.
		value = value[size:]
		// Decrypt it.
		stream := cipher.NewCTR(block, iv)
		stream.XORKeyStream(value, value)
		return value, nil
	}
	return nil, errors.New("securecookie: the value could not be decrypted")
}

// GenerateRandomKey creates a random key with the given strength.
func GenerateRandomKey(strength int) []byte {
	k := make([]byte, strength)
	if _, err := rand.Read(k); err != nil {
		return nil
	}
	return k
}