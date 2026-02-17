package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.sql.Timestamp;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.FavoritePhraseDto;
import com.example.FreStyle.entity.FavoritePhrase;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.FavoritePhraseRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetUserFavoritePhrasesUseCase テスト")
class GetUserFavoritePhrasesUseCaseTest {

    @Mock
    private FavoritePhraseRepository favoritePhraseRepository;

    @InjectMocks
    private GetUserFavoritePhrasesUseCase getUserFavoritePhrasesUseCase;

    @Test
    @DisplayName("ユーザーのお気に入りフレーズ一覧を取得できる")
    void execute_ReturnsFavoritePhrases() {
        User user = new User();
        user.setId(1);

        FavoritePhrase phrase1 = new FavoritePhrase();
        phrase1.setId(1);
        phrase1.setUser(user);
        phrase1.setOriginalText("確認お願いします");
        phrase1.setRephrasedText("ご確認いただけますでしょうか");
        phrase1.setPattern("フォーマル版");
        phrase1.setCreatedAt(new Timestamp(System.currentTimeMillis()));

        FavoritePhrase phrase2 = new FavoritePhrase();
        phrase2.setId(2);
        phrase2.setUser(user);
        phrase2.setOriginalText("すみません");
        phrase2.setRephrasedText("恐れ入りますが");
        phrase2.setPattern("ソフト版");
        phrase2.setCreatedAt(new Timestamp(System.currentTimeMillis()));

        when(favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(phrase1, phrase2));

        List<FavoritePhraseDto> result = getUserFavoritePhrasesUseCase.execute(1);

        assertEquals(2, result.size());
        assertEquals("確認お願いします", result.get(0).originalText());
        assertEquals("ご確認いただけますでしょうか", result.get(0).rephrasedText());
        assertEquals("フォーマル版", result.get(0).pattern());
    }

    @Test
    @DisplayName("お気に入りが空の場合は空リストを返す")
    void execute_ReturnsEmptyList() {
        when(favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());

        List<FavoritePhraseDto> result = getUserFavoritePhrasesUseCase.execute(1);

        assertTrue(result.isEmpty());
    }
}
