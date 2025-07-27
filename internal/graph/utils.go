package graph

import (
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func FindByAliasExact(entities []EntityType, term string) *EntityType {
	term = strings.ToLower(term)
	for _, e := range entities {
		for _, alias := range e.Aliases {
			if strings.Contains(term, strings.ToLower(alias)) {
				return &e
			}
		}
	}
	return nil
}

func FindByFuzzyAlias(entities []EntityType, term string) *EntityType {
	term = strings.ToLower(term)
	for _, e := range entities {
		for _, alias := range e.Aliases {
			if fuzzy.MatchNormalized(alias, term) {
				return &e
			}
		}
	}
	return nil
}
