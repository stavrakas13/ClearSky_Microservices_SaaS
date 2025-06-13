package services

// Helper functions for calculating grades based on question scores and weights.

import (
	"View_personal_grades/models"
)

// CalculateTotalGrade computes the weighted grade out of 10 scaled to the
// maximum mark defined by the exam's MarkScale.

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
