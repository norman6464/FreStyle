package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.FavoritePhraseDto;
import com.example.FreStyle.repository.FavoritePhraseRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetUserFavoritePhrasesUseCase {

    private final FavoritePhraseRepository favoritePhraseRepository;

    @Transactional(readOnly = true)
    public List<FavoritePhraseDto> execute(Integer userId) {
        return favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(userId).stream()
                .map(phrase -> new FavoritePhraseDto(
                        phrase.getId(),
                        phrase.getOriginalText(),
                        phrase.getRephrasedText(),
                        phrase.getPattern(),
                        phrase.getCreatedAt() != null ? phrase.getCreatedAt().toInstant().toString() : null))
                .collect(Collectors.toList());
    }
}
