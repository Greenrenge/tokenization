package tokenize

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func getPointer(s string) *string {
	return &s
}

func TestCreateCharFilterMapper(t *testing.T) {
	type args struct {
		excludes []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "it should exclude the string we put in the excludes array",
			args: args{
				excludes: []string{"ไม่", "⌘", " 👨‍👩‍👧"},
			},
			want: map[string]string{
				"ไม่มี":                 "มี",
				"this is a book  👨‍👩‍👧": "this is a book ",
				"this is a book⌘":       "this is a book",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := CreateCharFilterMapper(tt.args.excludes)
			for k, v := range tt.want {

				if got := mapper(k); !reflect.DeepEqual(got, v) {
					t.Errorf("CreateCharFilterMapper() = %v, want %v", got, v)
				}
			}
		})
	}
}

func TestCreateSplitterRegExp(t *testing.T) {
	type args struct {
		regex *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			name: "it should split the string by the sentence",
			args: args{
				regex: regexp.MustCompile(`\s+|\n|\t`),
			},
			want: map[string][]string{
				"ไม่มี นะ": []string{"ไม่มี", "นะ"},
				`this an  
				thaist   ing`: []string{"this", "an", "thaist", "ing"},
				"thai	  land": []string{"thai", "land"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := CreateSplitterRegExp(tt.args.regex)
			for k, v := range tt.want {
				if got := splitter(k); !reflect.DeepEqual(got, v) {
					t.Errorf("CreateSplitterRegExp() = %v, want %v", got, v)
				}
			}
		})
	}
}

func TestCreateSplitterByRunes(t *testing.T) {
	type args struct {
		splitters []rune
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			name: "it should split the string by the sentence",
			args: args{
				splitters: []rune{' ', '\t', '\n'},
			},
			want: map[string][]string{
				"ไม่มี นะ": []string{"ไม่มี", "นะ"},
				`this an  
					thaist   ing`: []string{"this", "an", "thaist", "ing"},
				"thai	  land": []string{"thai", "land"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splitter := CreateSplitterByRunes(tt.args.splitters)
			for k, v := range tt.want {
				if got := splitter(k); !reflect.DeepEqual(got, v) {
					t.Errorf("CreateSplitterByRunes() = %v, want %v", got, v)
				}
			}
		})
	}
}

func TestCreateTokenization(t *testing.T) {
	type args struct {
		tokenFn  func(string) []string
		stemmer  map[string]string
		filterFn func(string) bool
	}

	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			name: "it make a tokenization process correctly",
			args: args{
				tokenFn: func(s string) []string {
					return strings.Split(s, "และ")
				},
				stemmer: map[string]string{"ร๊าก": "รัก", "thailand": "ไทย"},
				filterFn: func(s string) bool {
					if s == "เหี้ย" {
						return false
					}
					return true
				},
			},
			want: map[string][]string{
				"มีกันและกัน":                                    []string{"มีกัน", "กัน"},
				`ร๊ากและจะร๊าก กับรักและThailandและthaistและing`: []string{"รัก", "จะร๊าก กับรัก", `ไทย`, "thaist", "ing"},
				"เหี้ยและเหี้ย":                                  []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenization := CreateTokenization(tt.args.tokenFn, tt.args.stemmer, tt.args.filterFn)
			for k, v := range tt.want {
				if got := tokenization(k); !reflect.DeepEqual(got, v) {
					if len(got) != 0 || len(v) != 0 {
						t.Errorf("CreateTokenization() = %v, want %v", got, v)
					}
				}
			}
		})
	}
}

func TestCreateNGram(t *testing.T) {
	str := []string{"A", "B", "C", "D", "E", "F"}
	type args struct {
		gramConfig NGramConfig
		strIn      chan []string
		gramOut    chan string
	}
	tests := []struct {
		name        string
		args        args
		gramsResult []string
	}{
		{
			name: "it should create the correct gram 1-4",
			args: args{
				gramConfig: NGramConfig{
					MinGram:    1,
					MaxGram:    4,
					GramFilter: func(s string) bool { return true },
				},
				strIn:   make(chan []string),
				gramOut: make(chan string),
			},
			gramsResult: []string{"A", "AB", "ABC", "ABCD", "B", "BC", "BCD", "BCDE", "C", "CD", "CDE", "CDEF", "D", "DE", "DEF", "E", "EF", "F"},
		},
		{
			name: "it should create the correct 2-2",
			args: args{
				gramConfig: NGramConfig{
					MinGram:    2,
					MaxGram:    2,
					GramFilter: func(s string) bool { return true },
				},
				strIn:   make(chan []string),
				gramOut: make(chan string),
			},
			gramsResult: []string{"AB", "BC", "CD", "DE", "EF"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go CreateNGram(tt.args.strIn, tt.args.gramOut, tt.args.gramConfig)
			tt.args.strIn <- str
			close(tt.args.strIn)
			i := 0
			for gram := range tt.args.gramOut {
				if gram != tt.gramsResult[i] {
					t.Errorf("TestCreateNGram() = %v, want %v", gram, tt.gramsResult[i])
				}
				i++
			}
		})
	}
}
