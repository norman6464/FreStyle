package com.example.FreStyle.config;

import org.springframework.context.annotation.Configuration;
import org.springframework.messaging.simp.config.MessageBrokerRegistry;
import org.springframework.web.socket.config.annotation.EnableWebSocketMessageBroker;
import org.springframework.web.socket.config.annotation.StompEndpointRegistry;
import org.springframework.web.socket.config.annotation.WebSocketMessageBrokerConfigurer;

import lombok.RequiredArgsConstructor;

@Configuration
@EnableWebSocketMessageBroker
@RequiredArgsConstructor
public class WebSocketConfig implements WebSocketMessageBrokerConfigurer {

    private final WebSocketAuthHandshakeInterceptor authHandshakeInterceptor;

    @Override
    public void configureMessageBroker(MessageBrokerRegistry config) {
        config.enableSimpleBroker("/topic"); // Redis連携前はここ
        config.setApplicationDestinationPrefixes("/app");
    }

    // REST CorsConfigと同じオリジンを使用
    private static final String[] ALLOWED_ORIGINS = {
        "http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com",
        "https://dcd3m6lwt0z8u.cloudfront.net",
        "http://localhost:5173",
        "https://normanblog.com",
        "http://normanblog.com"
    };

    @Override
    public void registerStompEndpoints(StompEndpointRegistry registry) {
        // ユーザーチャット用エンドポイント
        registry.addEndpoint("/ws/chat")
                .addInterceptors(authHandshakeInterceptor)
                .setAllowedOriginPatterns(ALLOWED_ORIGINS)
                .withSockJS();

        // AIチャット用エンドポイント
        registry.addEndpoint("/ws/ai-chat")
                .addInterceptors(authHandshakeInterceptor)
                .setAllowedOriginPatterns(ALLOWED_ORIGINS)
                .withSockJS();
    }
}
