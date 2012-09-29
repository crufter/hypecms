package speaker

type Speaker struct {
	NounToVerbs map[string]interface{}
	Has func(string, string) bool
}

func New(a func(string, string) bool, b map[string]interface{}) *Speaker {
	return &Speaker{b, a}
}

func (t *Speaker) IsNoun(a string) bool {
	_, has := t.NounToVerbs[a]
	return has
}

func (t *Speaker) NounHasVerb(noun, verb string) bool {
	return t.VerbLocation(noun, verb) != ""
}

func (t *Speaker) VerbLocation(noun, verb string) string {
	val, has := t.NounToVerbs[noun]
	if !has {
		return ""
	}
	for _, v := range val.([]interface{}) {
		if t.Has(v.(string), verb) {
			return v.(string)
		}
	}
	return ""
}