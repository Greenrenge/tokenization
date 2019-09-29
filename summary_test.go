package tokenize

import (
	"reflect"
	"testing"

	cedar "github.com/iohub/ahocorasick"
	. "gopkg.in/ahmetb/go-linq.v3"
)

type ResultCheck struct {
	word     string
	matchVal []GramCount
}

func TestBuildDict(t *testing.T) {
	type args struct {
		grams []GramCount
	}
	tests := []struct {
		name        string
		args        args
		resultCheck ResultCheck
	}{
		{
			name: "it should create the right aho corasick dict",
			args: args{
				grams: []GramCount{
					GramCount{K: "ทำไม", V: 20},
					GramCount{K: "ทำไมถึงเป็นแบบนี้", V: 18},
					GramCount{K: "ยากจัง ทำไมนะ", V: 20},
				},
			},
			resultCheck: ResultCheck{
				word: "ก็ทำไมถึงเป็นแบบนี้",
				matchVal: []GramCount{
					GramCount{K: "ทำไม", V: 20},
					GramCount{K: "ทำไมถึงเป็นแบบนี้", V: 18},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := BuildDict(tt.args.grams)
			resp := matcher.Match([]byte(tt.resultCheck.word))
			result := []GramCount{}

			for resp.HasNext() {
				r := []GramCount{}
				items := resp.NextMatchItem([]byte(tt.resultCheck.word))

				From(items).SelectT(func(i cedar.MatchToken) GramCount {
					return i.Value.(GramCount)
				}).SortT(func(a GramCount, b GramCount) bool {
					return b.V < a.V
				}).ToSlice(&r)

				result = append(result, r...)
			}
			ascSort := func(a GramCount, b GramCount) bool {
				return b.V > a.V
			}
			sortedResult := From(result).SortT(ascSort).Results()

			sortedExpect := From(tt.resultCheck.matchVal).SortT(ascSort).Results()

			if !reflect.DeepEqual(sortedExpect, sortedResult) {
				t.Errorf("expect %v but got %v", sortedExpect, sortedResult)
			}
		})
	}
}

func TestGramCount_FindSubset(t *testing.T) {
	type fields struct {
		K string
		V int
	}
	type args struct {
		m                *cedar.Matcher
		delSet           map[string]bool
		differentPercent float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "it should put the correct subset gram to the delSet map",
			fields: fields{K: "ทำไมถึงเป็นแบบนี้", V: 180},
			args: args{
				m: BuildDict([]GramCount{
					GramCount{K: "ทำไม", V: 170},       // in delSet
					GramCount{K: "เป็นแบบนี้", V: 200}, //in delSet
					GramCount{K: "เป็น", V: 20},
					GramCount{K: "ทำไมถึงเป็นแบบนี้", V: 180},
					GramCount{K: "ยากจัง ทำไมนะ", V: 20},
				}),
				delSet:           map[string]bool{},
				differentPercent: 0.80,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gram := &GramCount{
				K: tt.fields.K,
				V: tt.fields.V,
			}
			gram.FindSubset(tt.args.m, tt.args.delSet, tt.args.differentPercent)
			delSet := tt.args.delSet
			if len(delSet) != 2 {
				t.Errorf(" delSet length is not correct, expect %d but got %d, %v", 2, len(delSet), delSet)
			}
			if !delSet["ทำไม"] {
				t.Errorf(" delSet is not contain the expected exclude %s", "ทำไม")
			}
			if !delSet["เป็นแบบนี้"] {
				t.Errorf(" delSet is not contain the expected exclude %s", "เป็นแบบนี้")
			}
		})
	}
}

func TestGroupByGramVal(t *testing.T) {
	type args struct {
		grams []GramCount
	}
	tests := []struct {
		name string
		args args
		want map[int][]GramCount
	}{
		{
			name: "it should group the grams by its value correctly",
			args: args{
				grams: []GramCount{
					GramCount{K: "AAA", V: 1},
					GramCount{K: "BBB", V: 2},
					GramCount{K: "CCC", V: 3},
					GramCount{K: "DDD", V: 2},
					GramCount{K: "EEE", V: 1},
					GramCount{K: "FFF", V: 4},
					GramCount{K: "GGG", V: 4},
				},
			},
			want: map[int][]GramCount{
				1: []GramCount{
					GramCount{K: "AAA", V: 1},
					GramCount{K: "EEE", V: 1},
				},
				2: []GramCount{
					GramCount{K: "BBB", V: 2},
					GramCount{K: "DDD", V: 2},
				},
				3: []GramCount{
					GramCount{K: "CCC", V: 3},
				},
				4: []GramCount{
					GramCount{K: "FFF", V: 4},
					GramCount{K: "GGG", V: 4},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GroupByGramVal(tt.args.grams); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GroupByGramVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterSimilarity(t *testing.T) {
	type args struct {
		grams []GramCount
		ratio float64
	}
	tests := []struct {
		name string
		args args
		want []GramCount
	}{
		{
			name: "it should filter some too much similar out of the collections",
			args: args{
				grams: []GramCount{
					GramCount{
						K: "พลาดโพสต์ของเรา",
						V: 363,
					},
					GramCount{
						K: "ไม่พลาดโพสต์ของ",
						V: 363,
					},
					GramCount{
						K: "ได้ไม่พลาดโพสต์",
						V: 363,
					},
					GramCount{
						K: "พลาดแล้วหล่ะ",
						V: 363,
					},
				},
				ratio: 0.7,
			},
			want: []GramCount{
				GramCount{
					K: "ได้ไม่พลาดโพสต์",
					V: 363,
				},
				GramCount{
					K: "พลาดแล้วหล่ะ",
					V: 363,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilterSimilarity(tt.args.grams, tt.args.ratio); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterSimilarity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSummaryResult(t *testing.T) {
	type args struct {
		grams []GramCount
		size  int
	}
	tests := []struct {
		name string
		args args
		want Summary
	}{
		{
			name: "it should sort the []GramCount and return by its Value only top size sent as arg",
			args: args{
				size: 3,
				grams: []GramCount{
					GramCount{
						K: "AAA",
						V: 1,
					},
					GramCount{
						K: "BBB",
						V: 30,
					},
					GramCount{
						K: "CCC",
						V: 39,
					},
					GramCount{
						K: "DDD",
						V: 38,
					},
					GramCount{
						K: "EEE",
						V: 39,
					},
					GramCount{
						K: "FFF",
						V: 30,
					},
				},
			},
			want: Summary{
				grams: []GramCount{
					GramCount{
						K: "CCC",
						V: 39,
					},
					GramCount{
						K: "EEE",
						V: 39,
					},
					GramCount{
						K: "DDD",
						V: 38,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := SummaryResult(tt.args.grams, tt.args.size); !reflect.DeepEqual(gotS, tt.want) {
				t.Errorf("SummaryResult() = %v, want %v", gotS, tt.want)
			}
		})
	}
}

func TestSummary_MarshalJson(t *testing.T) {
	type fields struct {
		grams []GramCount
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "it should marshal to json correctly",
			fields: fields{
				grams: []GramCount{
					GramCount{
						K: "CCC",
						V: 39,
					},
					GramCount{
						K: "EEE",
						V: 39,
					},
					GramCount{
						K: "DDD",
						V: 38,
					},
				},
			},
			want: `[["CCC",39],["EEE",39],["DDD",38]]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Summary{
				grams: tt.fields.grams,
			}
			got := s.MarshalJson()
			if !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("Summary.MarshalJson() = %s, want %s", got, tt.want)
			}
		})
	}
}
