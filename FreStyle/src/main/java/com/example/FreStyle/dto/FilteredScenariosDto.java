package com.example.FreStyle.dto;

import java.util.List;

public record FilteredScenariosDto(
        List<PracticeScenarioDto> scenarios,
        int totalCount,
        List<String> availableDifficulties,
        List<String> availableCategories) {
}
