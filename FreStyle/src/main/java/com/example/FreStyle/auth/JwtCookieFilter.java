package com.example.FreStyle.auth;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.io.IOException;

// 1リクエスト1回のみ実行するクラスを継承している
@Component
public class JwtCookieFilter extends OncePerRequestFilter {

    // ← ここでLoggerを宣言
    private static final Logger log = LoggerFactory.getLogger(JwtCookieFilter.class);

    @Override
    protected void doFilterInternal(HttpServletRequest request,
                                    HttpServletResponse response,
                                    FilterChain filterChain)
            throws ServletException, IOException {

        String proto = request.getHeader("X-Forwarded-Proto");
        String forwardedFor = request.getHeader("X-Forwarded-For");

        // SLF4Jでログ出力
        log.info("Incoming request: method={}, uri={}, X-Forwarded-Proto={}, X-Forwarded-For={}",
                 request.getMethod(), request.getRequestURI(), proto, forwardedFor);

        HttpServletRequest wrappedRequest = request;

        if (request.getCookies() != null) {
            for (Cookie cookie : request.getCookies()) {
                if ("ACCESS_TOKEN".equals(cookie.getName())) {
                    String bearerToken = "Bearer " + cookie.getValue();
                    wrappedRequest = new HttpServletRequestWrapperWithHeader(request, "Authorization", bearerToken);
                    log.info("JwtCookieFilter: wrapped request with Authorization header");
                    break;
                }
            }
        }

        filterChain.doFilter(wrappedRequest, response);
    }
}