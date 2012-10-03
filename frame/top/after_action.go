package top

import(
	"net/url"
	"strings"
	"fmt"
	"net/http"
	"encoding/json"
)

// After running a background operation this either redirects with data in url paramters or prints out the json encoded result.
func (t *Top) actionResponse(err error, action_name string) {
	uni := t.uni
	if t.config.Debug {
		fmt.Println(uni.Req.Referer())
		fmt.Println("	", err)
	}
	_, is_json := uni.Req.Form["json"]
	redir := uni.Req.Referer()
	if red, ok := uni.Dat["redirect"]; ok {
		redir = red.(string)
	} else if post_red, okr := uni.Req.Form["redirect"]; okr && len(post_red) == 1 {
		redir = post_red[1]
	}
	var cont map[string]interface{}
	cont_i, has := uni.Dat["_cont"]
	if has {
		cont = cont_i.(map[string]interface{})
	} else {
		cont = map[string]interface{}{}
	}
	redir = appendParams(redir, action_name, err, cont)
	if is_json {
		cont["redirect"] = redir
		if err == nil {
			cont["ok"] = true
		} else {
			cont["error"] = err.Error()
		}
		var v []byte
		if _, fmt := uni.Req.Form["fmt"]; fmt {
			v, _ = json.MarshalIndent(cont, "", "    ")
		} else {
			v, _ = json.Marshal(cont)
		}
		uni.Put(string(v))
	} else {
		http.Redirect(uni.W, uni.Req, redir, 303)
	}
}

// This writes all necessary information after a background operation into the redirect url, and deletes
// parts which were when a previous background op ran.
func appendParams(url_str string, action_name string, err error, cont map[string]interface{}) string {
	p := strings.Split(url_str, "?")
	var inp string
	if len(p) > 1 {
		inp = p[1]
	} else {
		inp = ""
	}
	v, parserr := url.ParseQuery(inp)
	if parserr == nil {
		// Delete outdated information from url.
		for i := range v {
			if strings.HasPrefix(i, "-") {
				v.Del(i)
			}
		}
		// Write all data in cont into the url.
		for key, val := range cont {
			if key[0] == '!' {
				v.Set(key[1:], fmt.Sprint(val))
			} else {
				v.Set("-"+key, fmt.Sprint(val))
			}
		}
		v.Del("error")
		v.Del("ok") // See *1
		v.Del("action")
		if len(action_name) > 0 { // runDebug calls this function with an empty action name.
			v.Set("action", action_name)
		}
		if err == nil {
			v.Set("ok", "true") // This could be left out, but hey. *1
		} else {
			v.Set("error", err.Error())
		}
		quer := v.Encode()
		if len(quer) > 0 {
			return p[0] + "?" + quer
		}
	}
	return p[0]
}