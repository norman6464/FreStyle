package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.FavoritePhraseSummaryDto;
import com.example.FreStyle.entity.FavoritePhrase;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.FavoritePhraseRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetFavoritePhraseSummaryUseCase テスト")
class GetFavoritePhraseSummaryUseCaseTest {

    @Mock
    private FavoritePhraseRepository favoritePhraseRepository;

    @InjectMocks
    private GetFavoritePhraseSummaryUseCase getFavoritePhraseSummaryUseCase;

    private User user() {
        User u = new User();
        u.setId(1);
        return u;
    }

    private FavoritePhrase phrase(String pattern) {
        return new FavoritePhrase(null, user(), "original", "rephrased", pattern, new Timestamp(System.currentTimeMillis()));
    }

    @Test
    @DisplayName("パターン別のカウントを集計して返す")
    void execute_returnsPatternCounts() {
        List<FavoritePhrase> phrases = List.of(
                phrase("meeting"),
                phrase("meeting"),
                phrase("email"),
                phrase("meeting"),
                phrase("presentation")
        );

        when(favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(phrases);

        FavoritePhraseSummaryDto result = getFavoritePhraseSummaryUseCase.execute(1);

        assertThat(result.totalCount()).isEqualTo(5);
        assertThat(result.patternCounts()).hasSize(3);
        assertThat(result.patternCounts().stream()
                .filter(pc -> pc.pattern().equals("meeting"))
                .findFirst().get().count()).isEqualTo(3);
        verify(favoritePhraseRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("フレーズがない場合は空のサマリーを返す")
    void execute_emptyList_returnsZero() {
        when(favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(List.of());

        FavoritePhraseSummaryDto result = getFavoritePhraseSummaryUseCase.execute(1);

        assertThat(result.totalCount()).isZero();
        assertThat(result.patternCounts()).isEmpty();
        verify(favoritePhraseRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("全て同じパターンの場合")
    void execute_singlePattern_returnsOneEntry() {
        List<FavoritePhrase> phrases = List.of(
                phrase("email"),
                phrase("email"),
                phrase("email")
        );

        when(favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(phrases);

        FavoritePhraseSummaryDto result = getFavoritePhraseSummaryUseCase.execute(1);

        assertThat(result.totalCount()).isEqualTo(3);
        assertThat(result.patternCounts()).hasSize(1);
        assertThat(result.patternCounts().get(0).pattern()).isEqualTo("email");
        assertThat(result.patternCounts().get(0).count()).isEqualTo(3);
        verify(favoritePhraseRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("パターンカウントが降順でソートされる")
    void execute_sortedByCountDescending() {
        List<FavoritePhrase> phrases = List.of(
                phrase("email"),
                phrase("meeting"),
                phrase("meeting"),
                phrase("meeting"),
                phrase("email"),
                phrase("presentation")
        );

        when(favoritePhraseRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(phrases);

        FavoritePhraseSummaryDto result = getFavoritePhraseSummaryUseCase.execute(1);

        assertThat(result.totalCount()).isEqualTo(6);
        assertThat(result.patternCounts().get(0).pattern()).isEqualTo("meeting");
        assertThat(result.patternCounts().get(0).count()).isEqualTo(3);
        assertThat(result.patternCounts().get(1).count()).isEqualTo(2);
        assertThat(result.patternCounts().get(2).count()).isEqualTo(1);
        verify(favoritePhraseRepository).findByUserIdOrderByCreatedAtDesc(1);
    }
}
