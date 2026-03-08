package com.example.FreStyle.dto;

import java.util.List;

public record RankingDto(
    List<RankingEntryDto> entries,
    RankingEntryDto myRanking
) {
    public record RankingEntryDto(
        int rank,
        Integer userId,
        String username,
        String iconUrl,
        double averageScore,
        int sessionCount
    ) {}
}
