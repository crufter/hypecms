package display_model

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// Decides if a string should be localized.
const Min_loc_len = 8 // $loc.a.b
func IsLocString(s string) bool {
	return len(s) > Min_loc_len && string(s[0:4]) == "$loc." && strings.Index(s, ".") != -1
}

// Extracts the name of the localization file from the given loc string.
func ExtractLocName(s string) string {
	return strings.Split(s, ".")[1]
}

// TODO: This logic is very similar to what is being done in opesun/resolver. Check if a shared pattern could be extracted and reused.
func collect(i interface{}) []string {
	locfiles := []string{}
	switch val := i.(type) {
	case []interface{}:
		for _, v := range val {
			locfiles = append(locfiles, collect(v)...)
		}
	case map[string]interface{}:
		for _, v := range val {
			locfiles = append(locfiles, collect(v)...)
		}
	case string:
		if IsLocString(val) {
			locfiles = append(locfiles, ExtractLocName(val))
		}
	}
	return locfiles
}

// s absolute filepath
func locReader(s string) (map[string]interface{}, error) {
	file, err := ioutil.ReadFile(s)
	if err != nil {
		return nil, err
	}
	var v interface{}
	err = json.Unmarshal(file, &v)
	return v.(map[string]interface{}), err
}

// Extracts used multilingual variables from a template with regexp.
func CollectFromTempl(file_content string) map[string]struct{} {
	r := regexp.MustCompile(".loc.([a-zA-Z_.:/-])*")
	s := r.FindAllString(file_content, -1)
	c := map[string]struct{}{}
	for _, v := range s {
		spl := strings.Split(v, ".")
		if len(spl) > 3 {
			c[spl[2]] = struct{}{}
		}
	}
	return c
}

func CollectFromMap(dat map[string]interface{}) map[string]struct{} {
	sl := collect(dat)
	c := map[string]struct{}{}
	for _, v := range sl {
		c[v] = struct{}{}
	}
	return c
}

// Takes a list of localization filenames and tries to load every one of them, first from the template, then from the modules.
func ReadFiles(root, tplpath string, user_langs []string, locfiles map[string]struct{}, loc_reader func(s string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for i, _ := range locfiles {
		for _, lang := range user_langs {
			path := filepath.Join(root, tplpath, "loc", i+"."+lang)
			ma, err := loc_reader(path)
			if err == nil {
				ret[i] = ma
				break
			}
			path = filepath.Join(root, "modules", i, "tpl/loc", lang+".json")
			ma, err = loc_reader(path)
			if err == nil {
				ret[i] = ma
				break
			}
		}
	}
	return ret, nil
}

// tplpath is public/default or private/127.0.0.1/default
func LoadLocStrings(dat map[string]interface{}, user_langs []string, root, tplpath string, loc_reader func(s string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	if loc_reader == nil {
		loc_reader = locReader
	}
	locfiles := CollectFromMap(dat)
	return ReadFiles(root, tplpath, user_langs, locfiles, loc_reader)
}

// tplpath is public/default or private/127.0.0.1/default
func LoadLocTempl(file_content string, user_langs []string, root, tplpath string, loc_reader func(s string) (map[string]interface{}, error)) (map[string]interface{}, error) {
	if loc_reader == nil {
		loc_reader = locReader
	}
	locfiles := CollectFromTempl(file_content)
	return ReadFiles(root, tplpath, user_langs, locfiles, loc_reader)
}
