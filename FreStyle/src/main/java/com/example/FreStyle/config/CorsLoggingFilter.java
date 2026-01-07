package com.example.FreStyle.config;

import java.io.IOException;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

/**
 * CORSリクエストのデバッグ用ロギングフィルター
 * 全てのリクエストでCORS関連のヘッダーをログ出力します
 */
@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class CorsLoggingFilter implements Filter {

    private static final Logger logger = LoggerFactory.getLogger(CorsLoggingFilter.class);

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {
        
        HttpServletRequest httpRequest = (HttpServletRequest) request;
        HttpServletResponse httpResponse = (HttpServletResponse) response;
        
        String method = httpRequest.getMethod();
        String uri = httpRequest.getRequestURI();
        String origin = httpRequest.getHeader("Origin");
        String accessControlRequestMethod = httpRequest.getHeader("Access-Control-Request-Method");
        String accessControlRequestHeaders = httpRequest.getHeader("Access-Control-Request-Headers");
        
        // プリフライトリクエスト（OPTIONS）の場合は詳細ログ
        if ("OPTIONS".equalsIgnoreCase(method)) {
            logger.info("========== [CORS] プリフライトリクエスト検出 ==========");
            logger.info("[CORS] URI: {}", uri);
            logger.info("[CORS] Origin: {}", origin);
            logger.info("[CORS] Access-Control-Request-Method: {}", accessControlRequestMethod);
            logger.info("[CORS] Access-Control-Request-Headers: {}", accessControlRequestHeaders);
        } else if (origin != null) {
            // 通常のCORSリクエスト
            logger.info("[CORS] リクエスト - Method: {}, URI: {}, Origin: {}", method, uri, origin);
        }
        
        // フィルターチェーンを続行
        chain.doFilter(request, response);
        
        // レスポンスのCORSヘッダーをログ出力
        String allowOrigin = httpResponse.getHeader("Access-Control-Allow-Origin");
        String allowMethods = httpResponse.getHeader("Access-Control-Allow-Methods");
        String allowHeaders = httpResponse.getHeader("Access-Control-Allow-Headers");
        String allowCredentials = httpResponse.getHeader("Access-Control-Allow-Credentials");
        
        if ("OPTIONS".equalsIgnoreCase(method) || origin != null) {
            logger.info("[CORS] レスポンス - Status: {}", httpResponse.getStatus());
            logger.info("[CORS] Access-Control-Allow-Origin: {}", allowOrigin != null ? allowOrigin : "NOT SET ⚠️");
            logger.info("[CORS] Access-Control-Allow-Methods: {}", allowMethods != null ? allowMethods : "NOT SET");
            logger.info("[CORS] Access-Control-Allow-Headers: {}", allowHeaders != null ? allowHeaders : "NOT SET");
            logger.info("[CORS] Access-Control-Allow-Credentials: {}", allowCredentials != null ? allowCredentials : "NOT SET");
            
            // CORSエラーの可能性を警告
            if (origin != null && allowOrigin == null) {
                logger.error("========== [CORS] ⚠️ 警告: Access-Control-Allow-Origin が設定されていません！ ==========");
                logger.error("[CORS] リクエストOrigin: {} がCORS設定に含まれているか確認してください", origin);
            }
            
            if ("OPTIONS".equalsIgnoreCase(method)) {
                logger.info("========== [CORS] プリフライトリクエスト処理完了 ==========");
            }
        }
    }
}
