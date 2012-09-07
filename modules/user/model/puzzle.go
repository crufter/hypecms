package user_model

func Honeypot(inp map[string][]string, options map[string]interface{}) error {
	return nil
}

func Hashcash(inp map[string][]string, options map[string]interface{}) error {
	return nil
}

// Interpret puzzle group.
// Example puzzle_group:
// pg := []interface{}{"hashcash", "captcha", 1}
// This means "Solve hashcash and captcha. One of the puzzles may fail."
// This allows one to serve hashcash to those who have javascript enabled and fall back to captchas for those who don't.
func PuzzleGroup(puzzle_group []interface{}) (puzzles []string, can_fail int) {
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
