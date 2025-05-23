package services

import "stats_service/models"

func CalculateTotalGrade(scores []float64, weights []float64, scale models.MarkScale) float64 {
	if len(scores) != len(weights) || len(scores) == 0 {
		return 0.0
	}

	var totalWeight, weightedSum float64
	for i := range scores {
		normalized := scores[i] / 10.0
		weightedSum += normalized * weights[i]
		totalWeight += weights[i]
	}
	if totalWeight == 0 {
		return 0.0
	}
	return (weightedSum / totalWeight) * scale.Max
}
