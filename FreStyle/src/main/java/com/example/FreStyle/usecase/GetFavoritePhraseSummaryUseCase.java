package com.example.FreStyle.usecase;

import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FavoritePhraseSummaryDto;
import com.example.FreStyle.entity.FavoritePhrase;
import com.example.FreStyle.repository.FavoritePhraseRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetFavoritePhraseSummaryUseCase {

    private final FavoritePhraseRepository favoritePhraseRepository;

    @Transactional(readOnly = true)
    public FavoritePhraseSummaryDto execute(Integer userId) {
        List<FavoritePhrase> phrases = favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(userId);

        if (phrases.isEmpty()) {
            return new FavoritePhraseSummaryDto(0, List.of());
        }

        Map<String, Long> countByPattern = phrases.stream()
                .collect(Collectors.groupingBy(FavoritePhrase::getPattern, Collectors.counting()));

        List<FavoritePhraseSummaryDto.PatternCount> patternCounts = countByPattern.entrySet().stream()
                .sorted(Map.Entry.<String, Long>comparingByValue(Comparator.reverseOrder()))
                .map(e -> new FavoritePhraseSummaryDto.PatternCount(e.getKey(), e.getValue()))
                .toList();

        return new FavoritePhraseSummaryDto(phrases.size(), patternCounts);
    }
}
