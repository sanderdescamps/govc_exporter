package helper

import "strings"

type Matchable interface {
	Match(string) bool
}

type Matcher struct {
	Keywords []string
}

func (m *Matcher) First() string {
	if len(m.Keywords) > 0 {
		return m.Keywords[0]
	}
	return "none"
}

func NewMatcher(keywords ...string) *Matcher {
	return &Matcher{
		Keywords: keywords,
	}
}

func (m Matcher) Match(s string) bool {
	for _, keyword := range m.Keywords {
		if strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*" {
			return true
		}
	}
	return false
}

func (m Matcher) MatchAny(s ...string) bool {
	for _, keyword := range m.Keywords {
		for _, s := range s {
			if strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*" {
				return true
			}
		}
	}
	return false
}

func (m Matcher) MatchAll(s ...string) bool {
	for _, keyword := range m.Keywords {
		match := true
		for _, s := range s {
			if !(strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*") {
				match = false
			}
		}
		if match {
			return true
		}
	}
	return false
}
