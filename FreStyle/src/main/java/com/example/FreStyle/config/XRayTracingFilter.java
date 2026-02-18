package com.example.FreStyle.config;

import com.amazonaws.xray.AWSXRay;
import com.amazonaws.xray.entities.Segment;
import com.amazonaws.xray.entities.TraceHeader;

import jakarta.servlet.Filter;
import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.ServletRequest;
import jakarta.servlet.ServletResponse;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

import lombok.extern.slf4j.Slf4j;

import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.Map;
import java.util.Optional;

@Slf4j
@Component
@Order(Ordered.HIGHEST_PRECEDENCE + 1)
public class XRayTracingFilter implements Filter {

    private static final String TRACE_HEADER_KEY = "X-Amzn-Trace-Id";
    private static final String SEGMENT_NAME = "FreStyle";

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {

        if (!(request instanceof HttpServletRequest httpRequest)
                || !(response instanceof HttpServletResponse httpResponse)) {
            chain.doFilter(request, response);
            return;
        }

        String uri = httpRequest.getRequestURI();

        // ヘルスチェックとActuatorをトレース対象外にする
        if (uri.startsWith("/actuator")) {
            chain.doFilter(request, response);
            return;
        }

        Segment segment = beginSegment(httpRequest);

        try {
            chain.doFilter(request, response);

            int statusCode = httpResponse.getStatus();
            segment.putHttp("response", Map.of("status", statusCode));
            if (statusCode == 429) {
                segment.setThrottle(true);
            }
            if (statusCode >= 400 && statusCode < 500) {
                segment.setError(true);
            }
            if (statusCode >= 500) {
                segment.setFault(true);
            }
        } catch (Exception e) {
            segment.addException(e);
            segment.setFault(true);
            throw e;
        } finally {
            TraceHeader responseHeader = new TraceHeader(segment.getTraceId());
            httpResponse.addHeader(TRACE_HEADER_KEY, responseHeader.toString());
            AWSXRay.endSegment();
        }
    }

    private Segment beginSegment(HttpServletRequest request) {
        String traceHeaderValue = request.getHeader(TRACE_HEADER_KEY);

        Segment segment;
        if (traceHeaderValue != null) {
            TraceHeader incomingHeader = TraceHeader.fromString(traceHeaderValue);
            segment = AWSXRay.beginSegment(SEGMENT_NAME, incomingHeader.getRootTraceId(),
                    incomingHeader.getParentId());
        } else {
            segment = AWSXRay.beginSegment(SEGMENT_NAME);
        }

        segment.putHttp("request", Map.of(
                "url", request.getRequestURL().toString(),
                "method", request.getMethod(),
                "user_agent", Optional.ofNullable(request.getHeader("User-Agent")).orElse("unknown"),
                "client_ip", Optional.ofNullable(request.getHeader("X-Forwarded-For"))
                        .orElse(request.getRemoteAddr())
        ));

        return segment;
    }
}
