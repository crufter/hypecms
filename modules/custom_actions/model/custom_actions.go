package custom_actions_model

import(
	"labix.org/v2/mgo/bson"
	"github.com/opesun/hypecms/model/patterns"
	"github.com/opesun/extract"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"fmt"
)

type m map[string]interface{}

func isLegalOption(vote_options []string, input_vote string) bool {
	in := false
	for _, v := range vote_options {
		if v == input_vote {
			in = true
		}
	}
	return in
}

func generateVoteQuery(doc_id, user_id bson.ObjectId, vote_options []string) map[string]interface{} {
	q := map[string]interface{}{"_id": doc_id}
	and := []interface{}{}
	for _, v := range vote_options {
		r := m{
			v: m{
				"$ne": user_id,
			},
		}
		and = append(and, r)
	}
	q["$and"] = and
	return q
}

// Vote options are mutually exclusive.
// Example:
// {
//		"type": "vote"
//		"c": "contents",
//		"vote_options": ["like", "dislike"]
// }
func Vote(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	collection := action["c"].(string)
	vote_options		:= jsonp.ToStringSlice(action["vote_options"].([]interface{}))
	rules := map[string]interface{}{
		"document_id": 		"must",
		"vote_option":		"must",
	}
	dat, err := extract.New(rules).Extract(inp)
	if err != nil { return err }
	input_vote := dat["vote_option"].(string)
	if !isLegalOption(vote_options, input_vote) { return fmt.Errorf("Not a legal option.") }
	doc_id := patterns.ToIdWithCare(dat["document_id"].(string))
	user_id := user["_id"].(bson.ObjectId)
	q := generateVoteQuery(doc_id, user_id, vote_options)
	upd := m{
		"$addToSet": m{
			input_vote: user_id,
		},
		"$inc": m{
			input_vote + "_count": 1,
		},
	}
	return db.C(collection).Update(q, upd)
}

func RunAction(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string, action_name string) error {
	typ := action["type"].(string)
	var r error
	switch typ {
	case "vote":
		r = Vote(db, user, action, inp)
	default:
		r = fmt.Errorf("Unkown action %v at RunAction.", action_name)
	}
	return r
}