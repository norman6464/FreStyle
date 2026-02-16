package com.example.FreStyle.service;

import jakarta.servlet.http.HttpServletResponse;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AuthCookieServiceTest {

    @InjectMocks
    private AuthCookieService authCookieService;

    @Test
    @DisplayName("setAuthCookies: 4つのCookieをレスポンスに設定する")
    void setAuthCookiesSets4Cookies() {
        HttpServletResponse response = mock(HttpServletResponse.class);

        authCookieService.setAuthCookies(response, "access123", "refresh456", "test@example.com", "user1");

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(response, times(4)).addHeader(eq("Set-Cookie"), captor.capture());

        var cookies = captor.getAllValues();
        assertThat(cookies).anyMatch(c -> c.contains("ACCESS_TOKEN=access123"));
        assertThat(cookies).anyMatch(c -> c.contains("REFRESH_TOKEN=refresh456"));
        assertThat(cookies).anyMatch(c -> c.contains("EMAIL=test%40example.com") || c.contains("EMAIL=test@example.com"));
        assertThat(cookies).anyMatch(c -> c.contains("COGNITO_USERNAME=user1"));
    }

    @Test
    @DisplayName("clearRefreshTokenCookie: REFRESH_TOKENのCookieを削除する")
    void clearRefreshTokenCookieDeletesCookie() {
        HttpServletResponse response = mock(HttpServletResponse.class);

        authCookieService.clearRefreshTokenCookie(response);

        ArgumentCaptor<String> captor = ArgumentCaptor.forClass(String.class);
        verify(response).addHeader(eq("Set-Cookie"), captor.capture());

        assertThat(captor.getValue()).contains("REFRESH_TOKEN");
        assertThat(captor.getValue()).contains("Max-Age=0");
    }
}
