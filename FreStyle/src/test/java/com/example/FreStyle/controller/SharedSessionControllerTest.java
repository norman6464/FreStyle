package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.SharedSessionDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ShareSessionForm;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetPublicSessionsUseCase;
import com.example.FreStyle.usecase.ShareSessionUseCase;
import com.example.FreStyle.usecase.UnshareSessionUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("SharedSessionController")
class SharedSessionControllerTest {

    @Mock
    private GetPublicSessionsUseCase getPublicSessionsUseCase;

    @Mock
    private ShareSessionUseCase shareSessionUseCase;

    @Mock
    private UnshareSessionUseCase unshareSessionUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private SharedSessionController controller;

    private Jwt jwt;
    private User user;

    private void setUpAuth() {
        jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("test-sub");

        user = new User();
        user.setId(1);
        user.setName("テストユーザー");

        when(userIdentityService.findUserBySub("test-sub")).thenReturn(user);
    }

    @Test
    @DisplayName("公開セッション一覧を取得する")
    void shouldReturnPublicSessions() {
        SharedSessionDto dto = new SharedSessionDto(
                1, 10, "セッション", 1, "ユーザー", "https://icon.png", "説明", "2026-03-08T10:00:00");
        when(getPublicSessionsUseCase.execute()).thenReturn(List.of(dto));

        ResponseEntity<List<SharedSessionDto>> response = controller.getPublicSessions();

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        assertThat(response.getBody()).hasSize(1);
        assertThat(response.getBody().get(0).sessionTitle()).isEqualTo("セッション");
    }

    @Test
    @DisplayName("セッションを共有する")
    void shouldShareSession() {
        setUpAuth();
        ShareSessionForm form = new ShareSessionForm(10, "説明");
        SharedSessionDto dto = new SharedSessionDto(
                1, 10, "セッション", 1, "ユーザー", "https://icon.png", "説明", "2026-03-08T10:00:00");
        when(shareSessionUseCase.execute(eq(1), any(ShareSessionForm.class))).thenReturn(dto);

        ResponseEntity<SharedSessionDto> response = controller.shareSession(jwt, form);

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        assertThat(response.getBody().sessionId()).isEqualTo(10);
        verify(shareSessionUseCase).execute(1, form);
    }

    @Test
    @DisplayName("セッションの共有を解除する")
    void shouldUnshareSession() {
        setUpAuth();
        ResponseEntity<Void> response = controller.unshareSession(jwt, 10);

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        verify(unshareSessionUseCase).execute(1, 10);
    }
}
