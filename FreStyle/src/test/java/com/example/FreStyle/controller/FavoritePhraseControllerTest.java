package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.FavoritePhraseDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.AddFavoritePhraseUseCase;
import com.example.FreStyle.usecase.GetUserFavoritePhrasesUseCase;
import com.example.FreStyle.usecase.RemoveFavoritePhraseUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("FavoritePhraseController")
class FavoritePhraseControllerTest {

    @Mock
    private GetUserFavoritePhrasesUseCase getUserFavoritePhrasesUseCase;

    @Mock
    private AddFavoritePhraseUseCase addFavoritePhraseUseCase;

    @Mock
    private RemoveFavoritePhraseUseCase removeFavoritePhraseUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private FavoritePhraseController favoritePhraseController;

    private Jwt mockJwt;
    private User testUser;

    @BeforeEach
    void setUp() {
        mockJwt = mock(Jwt.class);
        when(mockJwt.getSubject()).thenReturn("sub-123");

        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");

        when(userIdentityService.findUserBySub("sub-123")).thenReturn(testUser);
    }

    @Nested
    @DisplayName("getFavoritePhrases")
    class GetFavoritePhrases {

        @Test
        @DisplayName("お気に入りフレーズ一覧を取得できる")
        void returnsPhrases() {
            FavoritePhraseDto dto = new FavoritePhraseDto(1, "元文", "変換文", "フォーマル版", "2026-01-01T00:00:00Z");
            when(getUserFavoritePhrasesUseCase.execute(1)).thenReturn(List.of(dto));

            ResponseEntity<List<FavoritePhraseDto>> response = favoritePhraseController.getFavoritePhrases(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).hasSize(1);
            assertThat(response.getBody().get(0).getOriginalText()).isEqualTo("元文");
        }
    }

    @Nested
    @DisplayName("addFavoritePhrase")
    class AddFavoritePhrase {

        @Test
        @DisplayName("お気に入りフレーズを追加できる")
        void addsPhrase() {
            Map<String, String> body = Map.of(
                    "originalText", "確認お願い",
                    "rephrasedText", "ご確認ください",
                    "pattern", "フォーマル版");

            ResponseEntity<Void> response = favoritePhraseController.addFavoritePhrase(mockJwt, body);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(addFavoritePhraseUseCase).execute(testUser, "確認お願い", "ご確認ください", "フォーマル版");
        }

        @Test
        @DisplayName("UseCase例外時にそのまま伝搬する")
        void propagatesException() {
            doThrow(new RuntimeException("追加失敗"))
                    .when(addFavoritePhraseUseCase).execute(testUser, "元文", "変換文", "パターン");

            org.junit.jupiter.api.Assertions.assertThrows(RuntimeException.class,
                    () -> favoritePhraseController.addFavoritePhrase(mockJwt,
                            Map.of("originalText", "元文", "rephrasedText", "変換文", "pattern", "パターン")));
        }
    }

    @Nested
    @DisplayName("removeFavoritePhrase")
    class RemoveFavoritePhrase {

        @Test
        @DisplayName("お気に入りフレーズを削除できる")
        void removesPhrase() {
            ResponseEntity<Void> response = favoritePhraseController.removeFavoritePhrase(mockJwt, 5);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(removeFavoritePhraseUseCase).execute(1, 5);
        }

        @Test
        @DisplayName("削除UseCase例外時にそのまま伝搬する")
        void throwsWhenUseCaseFails() {
            doThrow(new RuntimeException("削除失敗")).when(removeFavoritePhraseUseCase).execute(1, 99);

            org.junit.jupiter.api.Assertions.assertThrows(RuntimeException.class,
                    () -> favoritePhraseController.removeFavoritePhrase(mockJwt, 99));
        }
    }

    @Nested
    @DisplayName("共通エラー")
    class CommonErrors {

        @Test
        @DisplayName("UserIdentityService例外時にそのまま伝搬する")
        void throwsWhenUserNotFound() {
            when(userIdentityService.findUserBySub("sub-123"))
                    .thenThrow(new RuntimeException("ユーザーが見つかりません"));

            org.junit.jupiter.api.Assertions.assertThrows(RuntimeException.class,
                    () -> favoritePhraseController.getFavoritePhrases(mockJwt));
        }
    }

    @Nested
    @DisplayName("getFavoritePhrases エッジケース")
    class GetFavoritePhrasesEdgeCases {

        @Test
        @DisplayName("お気に入りが空の場合も200と空リストを返す")
        void returnsEmptyList() {
            when(getUserFavoritePhrasesUseCase.execute(1)).thenReturn(List.of());

            ResponseEntity<List<FavoritePhraseDto>> response = favoritePhraseController.getFavoritePhrases(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEmpty();
        }
    }
}
