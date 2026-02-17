package com.example.FreStyle.dto;

import java.util.List;

public record FavoritePhraseSummaryDto(
        int totalCount,
        List<PatternCount> patternCounts) {

    public record PatternCount(
            String pattern,
            long count) {
    }
}
