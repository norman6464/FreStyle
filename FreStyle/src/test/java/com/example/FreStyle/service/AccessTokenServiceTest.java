package com.example.FreStyle.service;

import com.example.FreStyle.entity.AccessToken;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AccessTokenRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AccessTokenServiceTest {

    @Mock
    private AccessTokenRepository accessTokenRepository;

    @InjectMocks
    private AccessTokenService accessTokenService;

    @Test
    @DisplayName("saveTokens: トークンを正しく保存する")
    void saveTokens_savesCorrectly() {
        User user = new User();
        user.setId(1);

        accessTokenService.saveTokens(user, "access123", "refresh456");

        ArgumentCaptor<AccessToken> captor = ArgumentCaptor.forClass(AccessToken.class);
        verify(accessTokenRepository).save(captor.capture());

        AccessToken saved = captor.getValue();
        assertEquals("access123", saved.getAccessToken());
        assertEquals("refresh456", saved.getRefreshToken());
        assertEquals(user, saved.getUser());
        assertFalse(saved.getRevoked());
    }

    @Test
    @DisplayName("findAccessTokenByRefreshToken: 正常系")
    void findByRefreshToken_returnsToken() {
        AccessToken token = new AccessToken();
        token.setRefreshToken("refresh123");
        when(accessTokenRepository.findByRefreshToken("refresh123"))
                .thenReturn(Optional.of(token));

        AccessToken result = accessTokenService.findAccessTokenByRefreshToken("refresh123");

        assertEquals("refresh123", result.getRefreshToken());
    }

    @Test
    @DisplayName("findAccessTokenByRefreshToken: 無効なトークンで例外")
    void findByRefreshToken_throwsWhenNotFound() {
        when(accessTokenRepository.findByRefreshToken("invalid"))
                .thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> accessTokenService.findAccessTokenByRefreshToken("invalid"));
        assertEquals("無効なリフレッシュトークンです。", ex.getMessage());
    }

    @Test
    @DisplayName("updateTokens: アクセストークンを更新する")
    void updateTokens_updatesAccessToken() {
        AccessToken token = new AccessToken();
        token.setAccessToken("old");

        accessTokenService.updateTokens(token, "newAccessToken");

        assertEquals("newAccessToken", token.getAccessToken());
        verify(accessTokenRepository).save(token);
    }
}
