package com.example.FreStyle.usecase;

import com.example.FreStyle.entity.AccessToken;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.AuthCookieService;
import com.example.FreStyle.service.CognitoAuthService;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class CognitoRefreshTokenUseCaseTest {

    @Mock private AccessTokenService accessTokenService;
    @Mock private CognitoAuthService cognitoAuthService;

    @InjectMocks
    private CognitoRefreshTokenUseCase cognitoRefreshTokenUseCase;

    @Test
    @DisplayName("リフレッシュトークン更新成功時にResult（新トークン情報）を返す")
    void returnsResultOnSuccess() {
        AccessToken accessTokenEntity = new AccessToken();

        when(accessTokenService.findAccessTokenByRefreshToken("refresh-token-123"))
                .thenReturn(accessTokenEntity);
        when(cognitoAuthService.refreshAccessToken("refresh-token-123", "testuser"))
                .thenReturn(Map.of("accessToken", "new-access-token"));

        CognitoRefreshTokenUseCase.Result result =
                cognitoRefreshTokenUseCase.execute("refresh-token-123", "testuser");

        assertThat(result.newAccessToken()).isEqualTo("new-access-token");

        verify(accessTokenService).updateTokens(accessTokenEntity, "new-access-token");
    }

    @Test
    @DisplayName("無効なリフレッシュトークンで例外をスローする")
    void throwsWhenRefreshTokenInvalid() {
        when(accessTokenService.findAccessTokenByRefreshToken("invalid-token"))
                .thenThrow(new RuntimeException("無効なリフレッシュトークンです。"));

        assertThatThrownBy(() -> cognitoRefreshTokenUseCase.execute("invalid-token", "testuser"))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("無効なリフレッシュトークンです。");

        verify(cognitoAuthService, never()).refreshAccessToken(any(), any());
    }

    @Test
    @DisplayName("Cognitoトークン更新失敗時に例外をスローする")
    void throwsWhenCognitoRefreshFails() {
        AccessToken accessTokenEntity = new AccessToken();

        when(accessTokenService.findAccessTokenByRefreshToken("refresh-token-123"))
                .thenReturn(accessTokenEntity);
        when(cognitoAuthService.refreshAccessToken("refresh-token-123", "testuser"))
                .thenThrow(new RuntimeException("リフレッシュトークンが無効です。"));

        assertThatThrownBy(() -> cognitoRefreshTokenUseCase.execute("refresh-token-123", "testuser"))
                .isInstanceOf(RuntimeException.class);

        verify(accessTokenService, never()).updateTokens(any(), any());
    }

    @Test
    @DisplayName("updateTokensに正しいアクセストークンが渡される")
    void passesCorrectAccessTokenToUpdate() {
        AccessToken accessTokenEntity = new AccessToken();

        when(accessTokenService.findAccessTokenByRefreshToken("rt-abc"))
                .thenReturn(accessTokenEntity);
        when(cognitoAuthService.refreshAccessToken("rt-abc", "user1"))
                .thenReturn(Map.of("accessToken", "at-xyz"));

        cognitoRefreshTokenUseCase.execute("rt-abc", "user1");

        verify(accessTokenService).updateTokens(accessTokenEntity, "at-xyz");
    }

    @Test
    @DisplayName("refreshAccessTokenに正しいリフレッシュトークンとユーザー名が渡される")
    void passesCorrectArgsToRefresh() {
        AccessToken accessTokenEntity = new AccessToken();

        when(accessTokenService.findAccessTokenByRefreshToken("my-refresh"))
                .thenReturn(accessTokenEntity);
        when(cognitoAuthService.refreshAccessToken("my-refresh", "user2"))
                .thenReturn(Map.of("accessToken", "new-at"));

        cognitoRefreshTokenUseCase.execute("my-refresh", "user2");

        verify(accessTokenService).findAccessTokenByRefreshToken("my-refresh");
        verify(cognitoAuthService).refreshAccessToken("my-refresh", "user2");
    }
}
