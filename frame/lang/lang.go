package lang

import(
	"net/url"
	"strconv"
	"strings"
	"fmt"
	"regexp"
	iface "github.com/opesun/hypecms/frame/interfaces"
)

func ToCodeStyle(a string) string {
	a = strings.Replace(a, "-", " ", -1)
	a = strings.Replace(a, "_", " ", -1)
	a = strings.Title(a)
	return strings.Replace(a, " ", "", -1)
}

func ToURLStyle(a string) string {
	return back(a, "-")
}

func back(a, sep string) string {
	r := regexp.MustCompile("([A-Z])")
	res := r.ReplaceAll([]byte(a), []byte(" $1"))
	spl := strings.Split(string(res), " ")
	for i := range spl {
		spl[i] = strings.ToLower(spl[i])
	}
	return strings.Join(spl, sep)
}

func ToFileStyle(a string) string {
	return back(a, "_")
}

type Route struct {
	checked			int
	Words			[]string
	Queries			[]url.Values
}

type URLEncoder struct {
	r *Route
	s *Sentence
}

func NewURLEncoder(r *Route, s *Sentence) *URLEncoder {
	return &URLEncoder{r, s}
}

func (u *URLEncoder) actionPath(action_name string) string {
	var words []string
	if u.s.Verb == "Get" {
		words = append(words, u.r.Words...)
	} else {
		words = append(words, u.r.Words[:len(u.r.Words)-1]...)
	}
	words = append(words, action_name)
	path := "/"+strings.Join(words, "/")
	return path
}

func (u *URLEncoder) Url(action_name string) string {
	path := u.actionPath(action_name)
	qu := u.r.encodeQueries().Encode()
	if len(qu) > 0 {
		path = path+qu
	}
	return qu
}

type Form struct {
	FilterFields	url.Values
	ActionPath		string
	KeyPrefix		string
}

func keyPrefix(action_path string) string {
	return strconv.Itoa(len(strings.Split(action_path, "/"))-2)
}

func (u *URLEncoder) Form(action_name string) *Form {
	f := &Form{}
	f.ActionPath = u.actionPath(action_name)
	f.KeyPrefix = keyPrefix(f.ActionPath)
	f.FilterFields = u.r.encodeQueries()
	return f
}

func (r *Route) encodeQueries() url.Values {
	u := url.Values{}
	for i, v := range r.Queries {
		for j, x := range v {
			var key string
			if i != 0 {
				key = key + strconv.Itoa(i)
			}
			key = key+j
			for _, z := range x {
				u.Add(key, z)
			}
		}
	}
	return u
}

func (r *Route) Get() string {
	r.checked++
	return r.Words[len(r.Words)-r.checked]
}

func (r *Route) Got() int {
	return r.checked
}

func (r *Route) DropOne() {
	r.Words = r.Words[:len(r.Words)-1]
	r.Queries = r.Queries[:len(r.Queries)-1]
}

func (r *Route) HasMorePair() bool {
	return len(r.Words)>=2+r.checked
}

func sortParams(q url.Values) map[int]url.Values {
	sorted := map[int]url.Values{}
	for i, v := range q {
		num, err := strconv.Atoi(string(i[0]))
		nummed := false
		if err == nil {
			nummed = true
		} else {
			num = 0
		}
		if nummed {
			i = i[1:]
		}
		if _, has := sorted[num]; !has {
			sorted[num] = url.Values{}
		}
		for _, x := range v {
			sorted[num].Add(i, x)
		}
	}
	return sorted
}

func nextIsId(current, next string) bool {
	//return next[1] == '-' && current[0] == next[0]
	return len(next) == 24
}

func extractId(next string) string {
	//return strings.Split(next, "-")[1]
	return next
}

// New		Post
// Edit		Put							
func InterpretRoute(p string, q url.Values) (*Route, error) {
	ps := strings.Split(p, "/")
	r := &Route{}
	r.Queries = []url.Values{}
	r.Words = []string{}
	if len(ps) < 1 {
		return r, fmt.Errorf("Wtf.")
	}
	ps = ps[1:]		// First one is empty string.
	sorted := sortParams(q)
	skipped := 0
	for i:=0;i<len(ps);i++ {
		v := ps[i]
		r.Words = append(r.Words, v)
		r.Queries = append(r.Queries, url.Values{})
		qi := len(r.Words)-1
		if len(ps) > i+1 {	// We are not at the end.
			next := ps[i+1]
			if nextIsId(v, next) {	// Id query in url., eg /users/u-fxARrttgFd34xdv7
				skipped++
				r.Queries[qi].Add("id", extractId(next))
				i++
				continue
			}
		}
		r.Queries[qi] = sorted[qi-skipped]
	}
	return r, nil
}

type Sentence struct{
	Noun, Verb, Redundant string
}

func Translate(r *Route, a iface.Speaker) *Sentence {
	s := &Sentence{}
	if len(r.Words) == 1 {
		s.Noun = r.Words[0]
		s.Verb = "Get"
		return s
	}
	unstable := r.Get()
	must_be_noun := r.Get()
	if a.IsNoun(unstable) {
		s.Verb = "Get"
		s.Noun = unstable
	} else if a.NounHasVerb(must_be_noun, ToCodeStyle(unstable)) {
		s.Verb = ToCodeStyle(unstable)
		s.Noun = must_be_noun
	} else {
		s.Redundant = unstable
		r.DropOne()
		s.Verb = "Get"
		s.Noun = must_be_noun
	}
	if !a.IsNoun(s.Noun) {
		panic(fmt.Sprintf("%v is not a noun.", s.Noun))
	}
	return s
}