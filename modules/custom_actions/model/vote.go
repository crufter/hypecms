package custom_actions_model

import (
	"fmt"
	"github.com/opesun/extract"
	"github.com/opesun/hypecms/model/patterns"
	"github.com/opesun/jsonp"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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

// Generates the part of the query which makes sure that the user have not voted yet.
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

// Most return values ever.
func sharedProc(action map[string]interface{}, inp map[string][]string) (map[string]interface{}, string, []string, string, error) {
	collection := action["c"].(string)
	vote_options := jsonp.ToStringSlice(action["vote_options"].([]interface{}))
	rules := map[string]interface{}{
		"document_id": "must",
		"vote_option": "must",
	}
	dat, err := extract.New(rules).Extract(inp)
	if err != nil {
		return nil, "", nil, "", err
	}
	input_vote := dat["vote_option"].(string)
	if !isLegalOption(vote_options, input_vote) {
		return nil, "", nil, "", fmt.Errorf("Not a legal option.")
	}
	return dat, collection, vote_options, input_vote, nil
}

// Vote options are mutually exclusive.
// Example:
// {
//		"type": 		"vote",
//		"c": 			"contents",
//		"doc_type":		"blog",					// Optional
//		"can_unvote":	true,					// Optional, handled as false if unset or set and false.
//		"vote_options": ["like", "dislike"]
// }
//
// Checks if the vote option sent from UI is legal.
// Makes sure user have not voted yet, pushes user_id into a field named equal to a member of "vote_options", and increments a field named equal to a member of "vote_options" + "_count".
func Vote(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	doc_type, has_dt := action["doc_typ"]
	dat, collection, vote_options, input_vote, err := sharedProc(action, inp)
	if err != nil {
		return err
	}
	doc_id := patterns.ToIdWithCare(dat["document_id"].(string))
	user_id := user["_id"].(bson.ObjectId)
	q := generateVoteQuery(doc_id, user_id, vote_options)
	if has_dt {
		q["type"] = doc_type.(string)
	}
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

// Checks if unvotes are approved at all. Returns error if not.
// Checks if vote option is legal.
// Checks if user indeed voted to that option. Return error if not.
// Decreases the counter field of the given vote option, and pulls the user_id from the field of the given vote option.
func Unvote(db *mgo.Database, user, action map[string]interface{}, inp map[string][]string) error {
	can_unvote, has_cu := action["can_unvote"]
	if !has_cu || can_unvote.(bool) == false {
		return fmt.Errorf("Can't unvote.")
	}
	dat, collection, _, input_vote, err := sharedProc(action, inp)
	if err != nil {
		return err
	}
	doc_id := patterns.ToIdWithCare(dat["document_id"].(string))
	user_id := user["_id"].(bson.ObjectId)
	q := m{"_id": doc_id, input_vote: user_id}
	count, err := db.C(collection).Find(q).Count()
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Can't unvote a doc which you haven't vote on yet.")
	}
	q = m{"_id": doc_id}
	upd := m{
		"$inc": m{
			input_vote + "_count": -1,
		},
		"$pull": m{
			input_vote: user_id,
		},
	}
	return db.C(collection).Update(q, upd)
}
