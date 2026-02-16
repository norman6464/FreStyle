package com.example.FreStyle.config;

import jakarta.servlet.FilterChain;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.mock.web.MockHttpServletRequest;
import org.springframework.mock.web.MockHttpServletResponse;

import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class CorsLoggingFilterTest {

    @InjectMocks
    private CorsLoggingFilter corsLoggingFilter;

    @Test
    @DisplayName("OPTIONSリクエストでフィルターチェーンが続行される")
    void doFilter_optionsRequestContinuesChain() throws Exception {
        MockHttpServletRequest request = new MockHttpServletRequest("OPTIONS", "/api/test");
        request.addHeader("Origin", "http://localhost:3000");
        request.addHeader("Access-Control-Request-Method", "POST");
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        corsLoggingFilter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("Originヘッダー付きリクエストでフィルターチェーンが続行される")
    void doFilter_originRequestContinuesChain() throws Exception {
        MockHttpServletRequest request = new MockHttpServletRequest("GET", "/api/test");
        request.addHeader("Origin", "http://localhost:3000");
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        corsLoggingFilter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("Originなしリクエストでもフィルターチェーンが続行される")
    void doFilter_noOriginRequestContinuesChain() throws Exception {
        MockHttpServletRequest request = new MockHttpServletRequest("GET", "/api/test");
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);

        corsLoggingFilter.doFilter(request, response, chain);

        verify(chain).doFilter(request, response);
    }

    @Test
    @DisplayName("フィルターチェーンが例外をスローした場合そのまま伝搬する")
    void doFilter_propagatesFilterChainException() throws Exception {
        MockHttpServletRequest request = new MockHttpServletRequest("POST", "/api/test");
        request.addHeader("Origin", "http://localhost:3000");
        MockHttpServletResponse response = new MockHttpServletResponse();
        FilterChain chain = mock(FilterChain.class);
        doThrow(new RuntimeException("チェーンエラー")).when(chain).doFilter(request, response);

        org.junit.jupiter.api.Assertions.assertThrows(RuntimeException.class,
                () -> corsLoggingFilter.doFilter(request, response, chain));
    }
}
