package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.FavoritePhrase;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.FavoritePhraseRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class AddFavoritePhraseUseCase {

    private final FavoritePhraseRepository favoritePhraseRepository;

    @Transactional
    public void execute(User user, String originalText, String rephrasedText, String pattern) {
        if (favoritePhraseRepository.existsByUserIdAndRephrasedTextAndPattern(
                user.getId(), rephrasedText, pattern)) {
            return;
        }

        FavoritePhrase phrase = new FavoritePhrase();
        phrase.setUser(user);
        phrase.setOriginalText(originalText);
        phrase.setRephrasedText(rephrasedText);
        phrase.setPattern(pattern);
        favoritePhraseRepository.save(phrase);
    }
}
