package display_model

import (
	"fmt"
	"github.com/opesun/jsonp"
	"github.com/opesun/require"
	"regexp"
)

const (
	beg = "{{load "
	end = "}}"
)

// Searches all {{load modname/filename.ext}} and replaces that with the proper requires then calls the require module on the file.
func Load(opt_i interface{}, root string, file []byte, get func(string, string) ([]byte, error)) ([]byte, error) {
	r := regexp.MustCompile(beg + "([a-zA-Z_.:/-])*" + end)
	s := r.FindAllString(string(file), -1)
	cut_beg := len(beg)
	cut_end := len(end)
	for _, v := range s {
		replacement := []byte{}
		load_name := v[cut_beg : len(v)-cut_end]
		loads, has := jsonp.Get(opt_i, load_name)
		if has {
			req_paths := jsonp.ToStringSlice(loads.([]interface{}))
			for _, x := range req_paths {
				replacement = append(replacement, []byte(fmt.Sprintf("{{require %v}}", x))...)
			}
		}
		if len(replacement) > 0 {
			str, err := require.RMem(root, replacement, get)
			if err != nil {
				return nil, err
			}
			replacement = []byte(str)
		}
		file = r.ReplaceAll(file, replacement)
	}
	return file, nil
}
