package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.FavoritePhraseRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("RemoveFavoritePhraseUseCase テスト")
class RemoveFavoritePhraseUseCaseTest {

    @Mock
    private FavoritePhraseRepository favoritePhraseRepository;

    @InjectMocks
    private RemoveFavoritePhraseUseCase removeFavoritePhraseUseCase;

    @Test
    @DisplayName("お気に入りフレーズを削除できる")
    void execute_RemovesPhrase() {
        removeFavoritePhraseUseCase.execute(1, 5);

        verify(favoritePhraseRepository).deleteByIdAndUserId(5, 1);
    }

    @Test
    @DisplayName("引数がrepositoryに正しい順序で渡される（phraseId, userId）")
    void execute_PassesArgumentsInCorrectOrder() {
        removeFavoritePhraseUseCase.execute(10, 20);

        verify(favoritePhraseRepository).deleteByIdAndUserId(20, 10);
    }

    @Test
    @DisplayName("repositoryが例外をスローした場合そのまま伝搬する")
    void execute_PropagatesRepositoryException() {
        doThrow(new RuntimeException("DB接続エラー"))
                .when(favoritePhraseRepository).deleteByIdAndUserId(5, 1);

        assertThatThrownBy(() -> removeFavoritePhraseUseCase.execute(1, 5))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("DB接続エラー");
    }
}
