package com.example.FreStyle.auth;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import lombok.extern.slf4j.Slf4j;
import java.io.IOException;

// 1リクエスト1回のみ実行するクラスを継承している
@Slf4j
@Component
public class JwtCookieFilter extends OncePerRequestFilter {

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

        // クッキーが存在するか確認
        if (request.getCookies() == null) {
            log.warn("JwtCookieFilter: No cookies found in request");
        } else {
            log.debug("JwtCookieFilter: Found {} cookies", request.getCookies().length);
            for (Cookie cookie : request.getCookies()) {
                log.debug("JwtCookieFilter: Cookie name={}, value={}", cookie.getName(), 
                         cookie.getValue() != null ? cookie.getValue().substring(0, Math.min(20, cookie.getValue().length())) + "..." : "null");
                
                if ("ACCESS_TOKEN".equals(cookie.getName())) {
                    String bearerToken = "Bearer " + cookie.getValue();
                    wrappedRequest = new HttpServletRequestWrapperWithHeader(request, "Authorization", bearerToken);
                    log.info("JwtCookieFilter: Successfully wrapped request with Authorization header from ACCESS_TOKEN cookie");
                    break;
                }
            }
            
            // ACCESS_TOKENが見つからない場合
            if (wrappedRequest == request) {
                log.warn("JwtCookieFilter: ACCESS_TOKEN cookie not found in cookies");
            }
        }

        long start = System.currentTimeMillis();
        filterChain.doFilter(wrappedRequest, response);
        long elapsed = System.currentTimeMillis() - start;
        if (elapsed > 100) {
            log.warn("[REQUEST-TIMING] {} {} - {}ms (status={})", request.getMethod(), request.getRequestURI(), elapsed, response.getStatus());
        } else {
            log.info("[REQUEST-TIMING] {} {} - {}ms (status={})", request.getMethod(), request.getRequestURI(), elapsed, response.getStatus());
        }
    }
}