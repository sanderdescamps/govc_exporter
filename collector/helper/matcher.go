package helper

import "strings"

type Matcher struct {
	Kewords []string
}

func (m *Matcher) First() string {
	if len(m.Kewords) > 0 {
		return m.Kewords[0]
	}
	return "none"
}

func NewMatcher(Kewords ...string) *Matcher {
	return &Matcher{
		Kewords: Kewords,
	}
}

func (m Matcher) Match(s string) bool {
	for _, keyword := range m.Kewords {
		if strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*" {
			return true
		}
	}
	return false
}

func (m Matcher) MatchAny(s ...string) bool {
	for _, keyword := range m.Kewords {
		for _, s := range s {
			if strings.EqualFold(keyword, s) || strings.ToLower(s) == "all" || s == "*" {
				return true
			}
		}
	}
	return false
}

func (m Matcher) MatchAll(s ...string) bool {
	for _, keyword := range m.Kewords {
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
