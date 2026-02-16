package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.repository.FavoritePhraseRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class RemoveFavoritePhraseUseCase {

    private final FavoritePhraseRepository favoritePhraseRepository;

    @Transactional
    public void execute(Integer userId, Integer phraseId) {
        favoritePhraseRepository.deleteByIdAndUserId(phraseId, userId);
    }
}
