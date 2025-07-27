package graph

import (
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

func FindBestFuzzyMatch(question string) string {
	entities := EntityTypesCache
	question = strings.ToLower(question)

	var bestEntity string
	bestScore := -1

	for _, entity := range entities {
		for _, alias := range entity.Aliases {
			score := fuzzy.RankMatchNormalizedFold(question, alias)
			if score > bestScore {
				bestScore = score
				bestEntity = entity.Name
			}
		}
	}

	if bestScore >= 60 {
		return bestEntity
	}
	return ""
}
