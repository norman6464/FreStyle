package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.AddScenarioBookmarkUseCase;
import com.example.FreStyle.usecase.GetUserBookmarksUseCase;
import com.example.FreStyle.usecase.RemoveScenarioBookmarkUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ScenarioBookmarkController")
class ScenarioBookmarkControllerTest {

    @Mock
    private GetUserBookmarksUseCase getUserBookmarksUseCase;

    @Mock
    private AddScenarioBookmarkUseCase addScenarioBookmarkUseCase;

    @Mock
    private RemoveScenarioBookmarkUseCase removeScenarioBookmarkUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private ScenarioBookmarkController controller;

    private Jwt jwt;
    private User user;

    @BeforeEach
    void setUp() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");

        user = new User();
        user.setId(1);
        user.setName("テストユーザー");

        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
    }

    @Nested
    @DisplayName("GET /api/bookmarks - ブックマーク一覧取得")
    class GetBookmarks {

        @Test
        @DisplayName("ブックマーク済みシナリオIDリストを返す")
        void shouldReturnBookmarkedIds() {
            when(getUserBookmarksUseCase.execute(1)).thenReturn(List.of(1, 3, 5));

            ResponseEntity<List<Integer>> response = controller.getBookmarks(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).containsExactly(1, 3, 5);
            verify(getUserBookmarksUseCase).execute(1);
        }

        @Test
        @DisplayName("ブックマークが無い場合は空リストを返す")
        void shouldReturnEmptyList() {
            when(getUserBookmarksUseCase.execute(1)).thenReturn(List.of());

            ResponseEntity<List<Integer>> response = controller.getBookmarks(jwt);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            assertThat(response.getBody()).isEmpty();
        }
    }

    @Nested
    @DisplayName("POST /api/bookmarks/{scenarioId} - ブックマーク追加")
    class AddBookmark {

        @Test
        @DisplayName("ブックマークを追加する")
        void shouldAddBookmark() {
            ResponseEntity<Void> response = controller.addBookmark(jwt, 3);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            verify(addScenarioBookmarkUseCase).execute(user, 3);
        }
    }

    @Nested
    @DisplayName("DELETE /api/bookmarks/{scenarioId} - ブックマーク削除")
    class RemoveBookmark {

        @Test
        @DisplayName("ブックマークを削除する")
        void shouldRemoveBookmark() {
            ResponseEntity<Void> response = controller.removeBookmark(jwt, 3);

            assertThat(response.getStatusCode().value()).isEqualTo(200);
            verify(removeScenarioBookmarkUseCase).execute(1, 3);
        }
    }
}
