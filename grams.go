package tokenize

import (
	"regexp"
	"strings"
)

// CreateCharFilterMapper : slice is a set of pointer, len so when slice len is changed inside the function, it would send &slice instead
func CreateCharFilterMapper(excludes []string) func(string) string {
	replacer := make([]string, 2*len(excludes))
	cLen := 0
	for _, exclude := range excludes {
		replacer[cLen*2] = exclude
		replacer[cLen*2+1] = ""
		cLen++
	}

	r := strings.NewReplacer(replacer...)
	return func(s string) string {
		s = r.Replace(s)
		return s
	}
}

// CreateSplitterRegExp is used (2) process to split text to a phrase, mainly use for thai lang
func CreateSplitterRegExp(regex *regexp.Regexp) func(string) []string {
	if regex == nil {
		return func(s string) []string {
			return []string{s}
		}
	}
	return func(s string) []string {
		arr := regex.Split(s, -1)
		return arr
	}
}

// CreateSplitterByRunes is used (2) process to split text to a phrase, mainly use for thai lang
func CreateSplitterByRunes(splitters []rune) func(string) []string {
	if len(splitters) == 0 {
		return func(s string) []string {
			return []string{s} //copy anyway
		}
	}
	return func(s string) []string {
		arr := strings.FieldsFunc(s, func(r rune) bool {
			for _, sp := range splitters {
				if sp == r {
					return true
				}
			}
			return false
		})
		return arr
	}
}

//routineable
func CreateTokenization(tokenFn func(string) []string, stemmer map[string]string, filterFn func(string) bool) func(string) []string {
	dict := stemmer
	//create token by tokenFn
	return func(s string) []string {
		tokens := tokenFn(s)
		var validTokens []string
		for _, token := range tokens {
			if !filterFn(token) {
				continue
			}
			if val := dict[strings.ToLower(token)]; val != "" {
				validTokens = append(validTokens, val)
			} else {
				validTokens = append(validTokens, token)
			}
		}
		return validTokens
	}
}

//select-case is wait for channels, like promise.race
//https://gobyexample.com/timeouts //timeout for channel
type NGramConfig struct {
	MinGram    int
	MaxGram    int
	GramFilter func(s string) bool
}

//routineable
func CreateNGram(strIn <-chan []string, gramOut chan<- string, gramConfig NGramConfig) {
	// pull []string from channel
	minGram := gramConfig.MinGram
	maxGram := gramConfig.MaxGram
	gramFilter := gramConfig.GramFilter

	for tokens := range strIn {
		for current := 0; current < len(tokens); current++ {
			var gram strings.Builder
			gram.WriteString(tokens[current])
			if minGram < 2 {
				gramOut <- gram.String()
			}
			for i := current + 1; i < len(tokens); i++ {
				gram.WriteString(tokens[i])
				if i-current+1 > maxGram {
					break
				}
				if i-current+1 >= minGram && gramFilter(gram.String()) {
					gramOut <- gram.String()
				}
			}

		}
	}
	close(gramOut)
}
