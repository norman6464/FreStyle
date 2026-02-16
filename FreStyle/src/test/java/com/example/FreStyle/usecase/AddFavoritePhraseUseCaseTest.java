package com.example.FreStyle.usecase;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.FavoritePhrase;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.FavoritePhraseRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("AddFavoritePhraseUseCase テスト")
class AddFavoritePhraseUseCaseTest {

    @Mock
    private FavoritePhraseRepository favoritePhraseRepository;

    @InjectMocks
    private AddFavoritePhraseUseCase addFavoritePhraseUseCase;

    private User testUser;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
    }

    @Test
    @DisplayName("新規お気に入りフレーズを追加できる")
    void execute_AddsNewPhrase() {
        when(favoritePhraseRepository.existsByUserIdAndRephrasedTextAndPattern(1, "ご確認ください", "フォーマル版"))
                .thenReturn(false);

        addFavoritePhraseUseCase.execute(testUser, "確認お願い", "ご確認ください", "フォーマル版");

        verify(favoritePhraseRepository).save(any(FavoritePhrase.class));
    }

    @Test
    @DisplayName("既に同一フレーズが存在する場合は重複保存しない")
    void execute_SkipsWhenAlreadyExists() {
        when(favoritePhraseRepository.existsByUserIdAndRephrasedTextAndPattern(1, "ご確認ください", "フォーマル版"))
                .thenReturn(true);

        addFavoritePhraseUseCase.execute(testUser, "確認お願い", "ご確認ください", "フォーマル版");

        verify(favoritePhraseRepository, never()).save(any(FavoritePhrase.class));
    }
}
