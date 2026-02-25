package com.example.FreStyle.config;

import java.util.Map;

import org.springframework.http.server.ServerHttpRequest;
import org.springframework.http.server.ServerHttpResponse;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.stereotype.Component;
import org.springframework.web.socket.WebSocketHandler;
import org.springframework.web.socket.server.HandshakeInterceptor;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Component
@RequiredArgsConstructor
@Slf4j
public class WebSocketAuthHandshakeInterceptor implements HandshakeInterceptor {

    public static final String AUTHENTICATED_USER_ID = "authenticatedUserId";

    private final UserIdentityService userIdentityService;

    @Override
    public boolean beforeHandshake(ServerHttpRequest request, ServerHttpResponse response,
                                    WebSocketHandler wsHandler, Map<String, Object> attributes) {
        try {
            Authentication authentication = SecurityContextHolder.getContext().getAuthentication();
            if (authentication != null && authentication.getPrincipal() instanceof Jwt jwt) {
                String sub = jwt.getSubject();
                User user = userIdentityService.findUserBySub(sub);
                attributes.put(AUTHENTICATED_USER_ID, user.getId());
                log.debug("WebSocket認証成功 - userId: {}", user.getId());
            } else {
                log.warn("WebSocketハンドシェイク: 認証情報なし");
            }
        } catch (Exception e) {
            log.warn("WebSocketハンドシェイク認証エラー: {}", e.getMessage());
        }
        return true;
    }

    @Override
    public void afterHandshake(ServerHttpRequest request, ServerHttpResponse response,
                               WebSocketHandler wsHandler, Exception exception) {
        // no-op
    }
}
