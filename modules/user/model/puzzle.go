package user_model

import (
	"fmt"
	"time"
	"strconv"
	"github.com/opesun/extract"
)

const(
	not_impl = "Not implemented yet."
	secret_salt = "xas_f9((!kcvm"
)

func SolveHoneypot(secret string, inp map[string][]string, puzzle_opts map[string]interface{}) error {
	return fmt.Errorf(not_impl)
}

func SolveHashcash(secret string, inp map[string][]string, puzzle_opts map[string]interface{}) error {
	return fmt.Errorf(not_impl)
}

func SolveTimer(secret string, inp map[string][]string, puzzle_opts map[string]interface{}) error {
	min_diff, ok := puzzle_opts["min_diff"].(int)
	if !ok {
		min_diff = 10
	}
	current := int(time.Now().Unix())
	r := map[string]interface{}{
		"__t": "must",
	}
	dat, err := extract.New(r).Extract(inp)
	if err != nil {
		return err
	}
	decrypted_v, err := decryptStr([]byte(secret_salt + secret), dat["__t"].(string))
	if err != nil {
		return err
	}
	stamp, err := strconv.Atoi(decrypted_v)
	if err != nil {
		return err
	}
	if current - stamp < min_diff {
		return fmt.Errorf("You submitted the form too quickly, wait %v seconds please.", min_diff)
	}
	return nil
}

// Interpret puzzle group.
// Example puzzle_group:
// pg := []interface{}{"hashcash", "captcha", 1}
// This means "Solve hashcash and captcha. One of the puzzles may fail."
// This allows one to serve hashcash to those who have javascript enabled and fall back to captchas for those who don't.
func InterpretPuzzleGroup(puzzle_group []interface{}) (puzzles []string, can_fail int) {
	puzzles = []string{}
	if len(puzzle_group) > 1 {
		num, is_num := puzzle_group[len(puzzle_group)-1].(int)
		if is_num { // Would make no sense otherwise.
			if num < len(puzzle_group) {
				can_fail = num
			}
			puzzle_group = puzzle_group[:len(puzzle_group)-1]
		}
	}
	for _, v := range puzzle_group {
		puzzles = append(puzzles, v.(string))
	}
	return
}

func ShowTimer(secret string, puzzle_opts map[string]interface{}) (string, error) {
	stamp := int(time.Now().Unix())
	encrypted_v, err := encryptStr([]byte(secret_salt + secret), strconv.Itoa(stamp))
	if err != nil {
		return "", err
	}
	// TODO: clear up this ugliness.
	return `<input name="__t" type="hidden" value="`+encrypted_v+`" />`, nil
}

func ShowHashcash(secret string, puzzle_opts map[string]interface{}) (string, error) {
	return "", fmt.Errorf(not_impl)
}

func ShowHoneypot(secret string, puzzle_opts map[string]interface{}) (string, error) {
	return "", fmt.Errorf(not_impl)
}
