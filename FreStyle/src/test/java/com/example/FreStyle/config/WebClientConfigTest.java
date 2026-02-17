package com.example.FreStyle.config;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.web.reactive.function.client.WebClient;

@DisplayName("WebClientConfig テスト")
class WebClientConfigTest {

    private WebClientConfig webClientConfig;

    @BeforeEach
    void setUp() {
        webClientConfig = new WebClientConfig();
    }

    @Test
    @DisplayName("WebClient Beanがnullでないことを確認する")
    void shouldReturnNonNullWebClient() {
        WebClient webClient = webClientConfig.webClient();

        assertThat(webClient).isNotNull();
    }

    @Test
    @DisplayName("WebClient BeanがWebClient型であることを確認する")
    void shouldReturnWebClientInstance() {
        WebClient webClient = webClientConfig.webClient();

        assertThat(webClient).isInstanceOf(WebClient.class);
    }
}
