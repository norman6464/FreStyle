package com.example.FreStyle.auth;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.io.IOException;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

class JwtCookieFilterTest {

    private JwtCookieFilter filter;
    private HttpServletRequest request;
    private HttpServletResponse response;
    private FilterChain filterChain;

    @BeforeEach
    void setUp() {
        filter = new JwtCookieFilter();
        request = mock(HttpServletRequest.class);
        response = mock(HttpServletResponse.class);
        filterChain = mock(FilterChain.class);
        when(request.getMethod()).thenReturn("GET");
        when(request.getRequestURI()).thenReturn("/api/test");
    }

    @Test
    void ACCESS_TOKENクッキーがある場合Authorizationヘッダーが付与される() throws ServletException, IOException {
        Cookie accessToken = new Cookie("ACCESS_TOKEN", "my-jwt-token");
        when(request.getCookies()).thenReturn(new Cookie[]{accessToken});

        filter.doFilterInternal(request, response, filterChain);

        ArgumentCaptor<HttpServletRequest> captor = ArgumentCaptor.forClass(HttpServletRequest.class);
        verify(filterChain).doFilter(captor.capture(), eq(response));

        HttpServletRequest wrappedRequest = captor.getValue();
        assertInstanceOf(HttpServletRequestWrapperWithHeader.class, wrappedRequest);
        assertEquals("Bearer my-jwt-token", wrappedRequest.getHeader("Authorization"));
    }

    @Test
    void クッキーがnullの場合リクエストがそのまま渡される() throws ServletException, IOException {
        when(request.getCookies()).thenReturn(null);

        filter.doFilterInternal(request, response, filterChain);

        verify(filterChain).doFilter(request, response);
    }

    @Test
    void ACCESS_TOKENクッキーがない場合リクエストがそのまま渡される() throws ServletException, IOException {
        Cookie otherCookie = new Cookie("OTHER_COOKIE", "value");
        when(request.getCookies()).thenReturn(new Cookie[]{otherCookie});

        filter.doFilterInternal(request, response, filterChain);

        verify(filterChain).doFilter(request, response);
    }

    @Test
    void 複数クッキーからACCESS_TOKENが正しく選択される() throws ServletException, IOException {
        Cookie session = new Cookie("SESSION_ID", "session-value");
        Cookie accessToken = new Cookie("ACCESS_TOKEN", "correct-token");
        Cookie refresh = new Cookie("REFRESH_TOKEN", "refresh-value");
        when(request.getCookies()).thenReturn(new Cookie[]{session, accessToken, refresh});

        filter.doFilterInternal(request, response, filterChain);

        ArgumentCaptor<HttpServletRequest> captor = ArgumentCaptor.forClass(HttpServletRequest.class);
        verify(filterChain).doFilter(captor.capture(), eq(response));

        assertEquals("Bearer correct-token", captor.getValue().getHeader("Authorization"));
    }
}
