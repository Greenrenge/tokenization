package tokenize

import (
	"encoding/json"
	"log"
	"math"

	"github.com/imjasonmiller/godice"
	"github.com/iohub/ahocorasick"
	"github.com/wesovilabs/koazee"
)

//already received from to redis
type GramCount struct {
	K string
	V int
}

// caller can do itself
//func FilterCheck(delSet map[string]bool)

// find if there is any grams are subset of this string,
// routineable --> wait group to done all then filter
func (gram *GramCount) FindSubset(m *cedar.Matcher, delSet map[string]bool, diffRatio float64) {
	// can be split to routines since we can ignore it since it's okay to save the same being deleted gram.
	if (delSet)[gram.K] {
		return
	}
	seq := []byte(gram.K)
	resp := m.Match(seq)
	for resp.HasNext() {
		subSetWords := resp.NextMatchItem(seq)
		for _, itr := range subSetWords {
			//key := m.Key(seq, itr)
			gramFound := itr.Value.(GramCount)
			//fmt.Printf("key:%s value:%v\n", key, gramFound)
			if gram.K != gramFound.K && (math.Abs(float64(gram.V-gramFound.V))/math.Max(float64(gram.V), float64(gramFound.V))) < diffRatio {
				delSet[gramFound.K] = true
			}
		}
	}
	// release buffer to sync.Pool
	resp.Release()
}

func BuildDict(grams []GramCount) *cedar.Matcher {
	m := cedar.NewMatcher()
	for _, gram := range grams {
		m.Insert([]byte(gram.K), gram)
	}
	m.Compile()
	return m
}

//groupByGramVal
func GroupByGramVal(grams []GramCount) map[int][]GramCount {
	stream := koazee.StreamOf(grams)
	grouped, _ := stream.GroupBy(func(g GramCount) int {
		return g.V
	})
	return grouped.Interface().(map[int][]GramCount)
}

//filterSimilarity --> routineable
//similarity := strsim.Compare("healed", "sealed")
func FilterSimilarity(grams []GramCount, ratio float64) []GramCount {
	//remove the first on we found the similar
	uniqueGrams := []GramCount{}
	for i, gram := range grams {
		isSimilarToOther := false
		for j := i + 1; j < len(grams); j++ {
			nextGram := grams[j]
			if godice.CompareString(gram.K, nextGram.K) > ratio {
				//similar
				isSimilarToOther = true
				break
			}
		}
		if !isSimilarToOther {
			uniqueGrams = append(uniqueGrams, gram)
		}
	}
	return uniqueGrams
}

type Summary struct {
	grams []GramCount
}

func (s *Summary) MarshalJson() []byte {
	forJSON := [][]interface{}{}

	for _, g := range s.grams {
		forJSON = append(forJSON, []interface{}{g.K, g.V})
	}
	var res []byte
	var err error
	if res, err = json.Marshal(forJSON); err != nil {
		log.Panicf("cannot marshal the Summary : %s", err.Error())
		return nil
	}
	return res

}

//sort the result
func SummaryResult(grams []GramCount, size int) (s Summary) {

	stream := koazee.StreamOf(grams).Sort(func(a GramCount, b GramCount) int {
		return b.V - a.V
	}).Take(0, size-1).Out().Val().([]GramCount)

	s = Summary{
		grams: stream,
	}
	return s
}

/**
type MsgCreateUserArray struct {
    CreateUser []MsgCreateUserJson `json:"array"`
}

type MsgCreateUserJson struct {
    EntityOrgName     string  `json:"entity_org_name"`
    EntityTitle       string  `json:"entity_title"`
    MsgBodyID         int64   `json:"msg_body_id,omitempty"`
    PosibbleUserEmail string  `json:"posibble_user_email"`
    PossibleUserName  string  `json:"possible_user_name"`
    UserPositionTitle string  `json:"user_position_title"`
}*/
