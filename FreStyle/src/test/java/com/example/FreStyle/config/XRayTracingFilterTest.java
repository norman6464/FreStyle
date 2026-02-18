package com.example.FreStyle.config;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.doThrow;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

import com.amazonaws.xray.AWSXRay;
import com.amazonaws.xray.AWSXRayRecorderBuilder;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;

@DisplayName("XRayTracingFilter テスト")
class XRayTracingFilterTest {

    private XRayTracingFilter filter;
    private MockHttpServletRequest request;
    private MockHttpServletResponse response;
    private FilterChain chain;

    @BeforeEach
    void setUp() {
        AWSXRay.setGlobalRecorder(AWSXRayRecorderBuilder.standard()
                .withContextMissingStrategy(new com.amazonaws.xray.strategy.LogErrorContextMissingStrategy())
                .build());
        filter = new XRayTracingFilter();
        request = new MockHttpServletRequest();
        response = new MockHttpServletResponse();
        chain = mock(FilterChain.class);
    }

    @AfterEach
    void tearDown() {
        AWSXRay.clearTraceEntity();
    }

    @Test
    @DisplayName("通常リクエストでフィルターチェーンが続行される")
    void normalRequestContinuesFilterChain() throws Exception {
        request.setRequestURI("/api/reports");
        request.setMethod("GET");

        filter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("Actuatorパスがトレース対象外になる")
    void actuatorPathSkipsTracing() throws Exception {
        request.setRequestURI("/actuator/health");
        request.setMethod("GET");

        filter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("レスポンスステータス4xxでerrorフラグが設定される")
    void status4xxSetsErrorFlag() throws Exception {
        request.setRequestURI("/api/test");
        request.setMethod("GET");
        response.setStatus(400);

        filter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("レスポンスステータス5xxでfaultフラグが設定される")
    void status5xxSetsFaultFlag() throws Exception {
        request.setRequestURI("/api/test");
        request.setMethod("GET");
        response.setStatus(500);

        filter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("レスポンスステータス429でthrottleフラグが設定される")
    void status429SetsThrottleFlag() throws Exception {
        request.setRequestURI("/api/test");
        request.setMethod("GET");
        response.setStatus(429);

        filter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("例外発生時にセグメントが正しく終了される")
    void exceptionEndsSegmentProperly() throws Exception {
        request.setRequestURI("/api/test");
        request.setMethod("GET");

        doThrow(new ServletException("テスト例外")).when(chain).doFilter(request, response);

        assertThatThrownBy(() -> filter.doFilter(request, response, chain))
                .isInstanceOf(ServletException.class)
                .hasMessage("テスト例外");
    }

    @Test
    @DisplayName("受信トレースヘッダーが伝搬される")
    void incomingTraceHeaderIsPropagated() throws Exception {
        request.setRequestURI("/api/test");
        request.setMethod("GET");
        request.addHeader("X-Amzn-Trace-Id", "Root=1-5f84c7a5-example;Parent=53995c3f42cd8ad8;Sampled=1");

        filter.doFilter(request, response, chain);

        assertThat(response.getHeader("X-Amzn-Trace-Id")).isNotNull();
        verify(chain).doFilter(request, response);
    }
}
