package com.example.FreStyle.dto;

import java.util.List;

public record RecommendedScenarioDto(
        List<ScenarioRecommendation> recommendations) {

    public record ScenarioRecommendation(
            Integer scenarioId,
            String scenarioName,
            String category,
            String difficulty,
            double averageScore,
            int practiceCount) {
    }
}
