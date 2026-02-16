package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.form.ProfileForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;

@ExtendWith(MockitoExtension.class)
@DisplayName("UpdateProfileUseCase テスト")
class UpdateProfileUseCaseTest {

    @Mock
    private CognitoAuthService cognitoAuthService;

    @Mock
    private UserService userService;

    @InjectMocks
    private UpdateProfileUseCase updateProfileUseCase;

    @Test
    @DisplayName("Cognitoユーザーの場合、CognitoとDBの両方を更新する")
    void execute_UpdatesCognitoAndDb_WhenCognitoUser() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-123");
        when(jwt.hasClaim("cognito:groups")).thenReturn(false);
        when(jwt.getTokenValue()).thenReturn("access-token");

        ProfileForm form = new ProfileForm("新しい名前", "新しい自己紹介", null);

        updateProfileUseCase.execute(jwt, form);

        verify(cognitoAuthService).updateUserProfile("access-token", "新しい名前");
        verify(userService).updateUser(form, "sub-123");
    }

    @Test
    @DisplayName("OIDCユーザーの場合、DBのみ更新する")
    void execute_UpdatesDbOnly_WhenOidcUser() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-456");
        when(jwt.hasClaim("cognito:groups")).thenReturn(true);

        ProfileForm form = new ProfileForm("OIDC名前", "OIDC自己紹介", null);

        updateProfileUseCase.execute(jwt, form);

        verify(cognitoAuthService, never()).updateUserProfile(anyString(), anyString());
        verify(userService).updateUser(form, "sub-456");
    }

    @Test
    @DisplayName("入力値不正時はIllegalArgumentExceptionをスロー")
    void execute_ThrowsException_WhenInvalidInput() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("sub-123");
        when(jwt.hasClaim("cognito:groups")).thenReturn(false);
        when(jwt.getTokenValue()).thenReturn("access-token");

        ProfileForm form = new ProfileForm("名前", "自己紹介", null);

        doThrow(new IllegalArgumentException("名前が不正です"))
                .when(cognitoAuthService).updateUserProfile("access-token", "名前");

        assertThrows(IllegalArgumentException.class, () -> {
            updateProfileUseCase.execute(jwt, form);
        });
    }
}
