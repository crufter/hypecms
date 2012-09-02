package user_model

import(
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"github.com/opesun/extract"
	"github.com/opesun/slugify"
	"github.com/opesun/hypecms/model/basic"
	ifaces "github.com/opesun/hypecms/interfaces"
	"crypto/sha1"
	"net/http"
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

func FindUser(db *mgo.Database, id interface{}) (map[string]interface{}, error) {
	v:= basic.Find(db, "users", id)
	if v != nil {
		delete(v, "password")
		return v, nil
	}
	return nil, fmt.Errorf("Can't find user with id %v.", id)
}

func namePass(db *mgo.Database, name, encoded_pass string) (map[string]interface{}, error) {
	var v interface{}
	err := db.C("users").Find(bson.M{"name": name, "password": encoded_pass}).One(&v)
	if err != nil {
		return nil, err
	}
	return basic.Convert(v).(map[string]interface{}), nil
}

// Everyone uses this to log in, admins, users, guest users and their mom.
func FindLogin(db *mgo.Database, inp map[string][]string) (map[string]interface{}, bson.ObjectId, error) {
	rule := map[string]interface{}{
		"name": 	"must",
		"password": "must",
	}
	d, err := extract.New(rule).Extract(inp)
	if err != nil {
		return nil, "", err
	}
	name := d["name"].(string)
	pass := EncodePass(d["password"].(string))
	user, err := namePass(db, name, pass)
	if err != nil {
		return nil, "", err
	}
	return user, user["_id"].(bson.ObjectId), nil
}

// Sets a cookie to w named "user" with a value of the encoded user_id.
// Admins, guests, registered users, everyone logs in with this.
func Login(w http.ResponseWriter, user_id bson.ObjectId, block_key []byte) error {
	block_key = block_key[0:block_size]
	if len(block_key) < block_size {
		return fmt.Errorf("Login: block_key length must be at least %v.", block_size)
	}
	id_b, err := encryptStr(block_key, user_id.Hex())
	encoded_id := string(id_b)
	if err != nil { return err }
	c := &http.Cookie{
		Name: "user",
		Value: encoded_id,
		MaxAge: 3600000,
		Path: "/",
	}
	http.SetCookie(w, c)
	return nil
}

func EmptyUser() map[string]interface{} {
	user := make(map[string]interface{})
	user["level"] = -1
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

// cookieval is encrypted
// Converts a string (a cookie) into an ObjectId.
func DecryptId(cookieval string, block_key []byte) (bson.ObjectId, error) {
	if len(block_key) < 16 {
		return "", fmt.Errorf("block_key length must be at least %v.", block_size)
	}
	if len(cookieval) < 0 {
		return "", fmt.Errorf("Nothing to decrypt.")
	}
	block_key = block_key[0:block_size]
	decr_id_b, err := decryptStr(block_key, cookieval)
	if err != nil { return "", err }
	return bson.ObjectIdHex(string(decr_id_b)), nil
}

// Builds a user from his Id and information in http_header.
func BuildUser(db *mgo.Database, ev ifaces.Event, user_id bson.ObjectId, http_header map[string][]string) (map[string]interface{}, error) {
	user, err := FindUser(db, user_id)
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

// This is public because the admin_model.RegUser needs it.
func EncodePass(pass string) string {
	h := sha1.New()
	io.WriteString(h, pass)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func NameAvailable(db *mgo.Database, name string) (bool, error) {
	var res []interface{}
	q := bson.M{"slug": slugify.S(name)}
	err := db.C("users").Find(q).All(&res)
	if err != nil { return false, err }
	if len(res) > 0 {
		return false, nil
	}
	return true, nil
}

func nameRule() map[string]interface{} {
	return map[string]interface{}{
		"type": "strings",
		"min":	4,
		"must":	true,
	}
}

// Set not existing fields.
// Checks if the they keys in b exists in a, and if not, sets a[b^i] = c^i
func setNE(a map[string]interface{}, b []string, c []interface{}) {
	if len(b) != len(c) { panic("b and c len must match.") }
	for i, v := range b {
		if _, has := a[v]; !has {
			a[v] = c[i]
		}
	}
}

func userDefaults(rules map[string]interface{}) {
	if rules == nil {
		rules = map[string]interface{}{}
	}
	name_rule := nameRule()
	pass_rule := map[string]interface{}{
		"type": "string",
		"min":	8,
		"must":	true,
	}
	b := []string{"name", "password", "password_again"}
	c := []interface{}{name_rule, pass_rule, pass_rule}
	setNE(rules, b, c)
}

// Registers a normal user with password and level 100.
// See RegisterGuest for an other kind of registration.
// See admin_model for registrations of admins.
func RegisterUser(db *mgo.Database, ev ifaces.Event, rules map[string]interface{}, inp map[string][]string) (bson.ObjectId, error) {
	userDefaults(rules)
	user, err := extract.New(rules).Extract(inp)
	if err != nil { return "", err }
	if user["password"].(string) != user["password_again"].(string) {
		return "", fmt.Errorf("Password and password confirmation differs.")
	}
	delete(user, "password_again")
	user["password"] = EncodePass(user["password"].(string))
	user["slug"] = slugify.S(user["slug"].(string))
	user["level"] = 100
	user_id := bson.NewObjectId()
	user["_id"] = user_id
	err = db.C("users").Insert(user)
	if err != nil {
		return "", fmt.Errorf("Name is not unique.")
	}
	delete(user, "password")
	ev.Trigger("user.register", user)
	return user_id, nil
}

func guestDefaults(rules map[string]interface{}) {
	if rules == nil {
		rules = map[string]interface{}{}
	}
	name_rule := nameRule()
	rules["name"] = name_rule
}

// Quickly register someone when he does an action as a guest.
// guest_rules can be nil.
func RegisterGuest(db *mgo.Database, ev ifaces.Event, guest_rules map[string]interface{}, inp map[string][]string) (bson.ObjectId, error) {
	guestDefaults(guest_rules)
	user, err := extract.New(guest_rules).Extract(inp)
	if err != nil { return "", err }
	user["level"] = 0
	user_id := bson.NewObjectId()
	user["_id"] = user_id
	err = db.C("users").Insert(user)
	if err != nil {
		return "", fmt.Errorf("Name is not unique.")
	}
	return user_id, nil
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

// The following functions are taken from securecookie package of the Gorilla web toolkit made by Rodrigo Moraes.
// Only modification was to make the GenerateRandomKey function private.

// encrypt encrypts a value using the given block in counter mode.
//
// A random initialization vector (http://goo.gl/zF67k) with the length of the
// block size is prepended to the resulting ciphertext.
func encrypt(block cipher.Block, value []byte) ([]byte, error) {
	iv := generateRandomKey(block.BlockSize())
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
func generateRandomKey(strength int) []byte {
	k := make([]byte, strength)
	if _, err := rand.Read(k); err != nil {
		return nil
	}
	return k
}