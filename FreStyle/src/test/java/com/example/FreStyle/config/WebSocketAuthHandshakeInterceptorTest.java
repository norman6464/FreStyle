package com.example.FreStyle.config;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.HashMap;
import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.server.ServerHttpRequest;
import org.springframework.http.server.ServerHttpResponse;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContext;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.socket.WebSocketHandler;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("WebSocketAuthHandshakeInterceptor")
class WebSocketAuthHandshakeInterceptorTest {

    @Mock private UserIdentityService userIdentityService;
    @Mock private ServerHttpRequest request;
    @Mock private ServerHttpResponse response;
    @Mock private WebSocketHandler wsHandler;

    @InjectMocks
    private WebSocketAuthHandshakeInterceptor interceptor;

    @Test
    @DisplayName("認証済みユーザーのIDがセッション属性に設定される")
    void authenticatedUser_setsUserIdInAttributes() throws Exception {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("cognito-sub-123");

        Authentication auth = mock(Authentication.class);
        when(auth.getPrincipal()).thenReturn(jwt);

        SecurityContext securityContext = mock(SecurityContext.class);
        when(securityContext.getAuthentication()).thenReturn(auth);
        SecurityContextHolder.setContext(securityContext);

        User user = new User();
        user.setId(42);
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(user);

        Map<String, Object> attributes = new HashMap<>();
        boolean result = interceptor.beforeHandshake(request, response, wsHandler, attributes);

        assertTrue(result);
        assertEquals(42, attributes.get(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID));

        SecurityContextHolder.clearContext();
    }

    @Test
    @DisplayName("認証情報がない場合はユーザーIDが設定されない")
    void noAuthentication_doesNotSetUserId() throws Exception {
        SecurityContext securityContext = mock(SecurityContext.class);
        when(securityContext.getAuthentication()).thenReturn(null);
        SecurityContextHolder.setContext(securityContext);

        Map<String, Object> attributes = new HashMap<>();
        boolean result = interceptor.beforeHandshake(request, response, wsHandler, attributes);

        assertTrue(result);
        assertNull(attributes.get(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID));

        SecurityContextHolder.clearContext();
    }

    @Test
    @DisplayName("UserIdentityServiceが例外をスローしても接続を許可する")
    void exceptionDuringLookup_stillAllowsConnection() throws Exception {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("unknown-sub");

        Authentication auth = mock(Authentication.class);
        when(auth.getPrincipal()).thenReturn(jwt);

        SecurityContext securityContext = mock(SecurityContext.class);
        when(securityContext.getAuthentication()).thenReturn(auth);
        SecurityContextHolder.setContext(securityContext);

        when(userIdentityService.findUserBySub("unknown-sub")).thenThrow(new RuntimeException("ユーザー不明"));

        Map<String, Object> attributes = new HashMap<>();
        boolean result = interceptor.beforeHandshake(request, response, wsHandler, attributes);

        assertTrue(result);
        assertNull(attributes.get(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID));

        SecurityContextHolder.clearContext();
    }
}
