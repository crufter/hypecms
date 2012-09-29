package lang

import(
	"net/url"
	"strconv"
	"strings"
	"fmt"
	iface "github.com/opesun/hypecms/frame/interfaces"
)

func sanitize(a string) string {
	a = strings.Replace(a, "-", " ", -1)
	a = strings.Replace(a, "_", " ", -1)
	a = strings.Title(a)
	return strings.Replace(a, " ", "", -1)
}

type Route struct {
	checked			int
	Words			[]string
	Queries			[]url.Values
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
			if next[1] == '-' && next[0] == v[0] {	// Id query in url., eg /users/u-fxARrttgFd34xdv7
				skipped++
				r.Queries[qi].Add("id", strings.Split(next, "-")[1])
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
	} else if a.NounHasVerb(must_be_noun, sanitize(unstable)) {
		s.Verb = sanitize(unstable)
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