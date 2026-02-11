package com.example.FreStyle.config;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.Signature;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

class RequestTimingAspectTest {

    private RequestTimingAspect aspect;
    private MeterRegistry meterRegistry;

    @BeforeEach
    void setUp() {
        meterRegistry = new SimpleMeterRegistry();
        aspect = new RequestTimingAspect(meterRegistry);
    }

    private ProceedingJoinPoint createMockJoinPoint(String className, String methodName, Object returnValue) throws Throwable {
        ProceedingJoinPoint joinPoint = mock(ProceedingJoinPoint.class);
        Signature signature = mock(Signature.class);
        when(signature.getDeclaringType()).thenReturn((Class) Object.class);
        when(signature.getDeclaringTypeName()).thenReturn(className);
        when(signature.getName()).thenReturn(methodName);
        when(joinPoint.getSignature()).thenReturn(signature);
        when(joinPoint.proceed()).thenReturn(returnValue);
        return joinPoint;
    }

    @Nested
    @DisplayName("timeControllerMethods - コントローラーメソッドの計測")
    class TimeControllerMethodsTest {

        @Test
        @DisplayName("コントローラーメソッドの処理時間がタイマーに記録される")
        void shouldRecordControllerMethodTiming() throws Throwable {
            ProceedingJoinPoint joinPoint = createMockJoinPoint(
                    "com.example.FreStyle.controller.UserController", "getUsers", "result");

            Object result = aspect.timeControllerMethods(joinPoint);

            assertThat(result).isEqualTo("result");
            Timer timer = meterRegistry.find("method.timed")
                    .tag("class", "com.example.FreStyle.controller.UserController")
                    .tag("method", "getUsers")
                    .tag("layer", "controller")
                    .timer();
            assertThat(timer).isNotNull();
            assertThat(timer.count()).isEqualTo(1);
        }

        @Test
        @DisplayName("元のメソッドが呼び出される")
        void shouldCallOriginalMethod() throws Throwable {
            ProceedingJoinPoint joinPoint = createMockJoinPoint(
                    "com.example.FreStyle.controller.TestController", "test", null);

            aspect.timeControllerMethods(joinPoint);

            verify(joinPoint).proceed();
        }
    }

    @Nested
    @DisplayName("timeServiceMethods - サービスメソッドの計測")
    class TimeServiceMethodsTest {

        @Test
        @DisplayName("サービスメソッドの処理時間がタイマーに記録される")
        void shouldRecordServiceMethodTiming() throws Throwable {
            ProceedingJoinPoint joinPoint = createMockJoinPoint(
                    "com.example.FreStyle.service.ChatService", "sendMessage", "ok");

            Object result = aspect.timeServiceMethods(joinPoint);

            assertThat(result).isEqualTo("ok");
            Timer timer = meterRegistry.find("method.timed")
                    .tag("class", "com.example.FreStyle.service.ChatService")
                    .tag("method", "sendMessage")
                    .tag("layer", "service")
                    .timer();
            assertThat(timer).isNotNull();
            assertThat(timer.count()).isEqualTo(1);
        }

        @Test
        @DisplayName("例外発生時もタイマーが記録される")
        void shouldRecordTimingEvenOnException() throws Throwable {
            ProceedingJoinPoint joinPoint = createMockJoinPoint(
                    "com.example.FreStyle.service.ErrorService", "fail", null);
            when(joinPoint.proceed()).thenThrow(new RuntimeException("test error"));

            try {
                aspect.timeServiceMethods(joinPoint);
            } catch (RuntimeException ignored) {
            }

            Timer timer = meterRegistry.find("method.timed")
                    .tag("class", "com.example.FreStyle.service.ErrorService")
                    .tag("method", "fail")
                    .tag("layer", "service")
                    .timer();
            assertThat(timer).isNotNull();
            assertThat(timer.count()).isEqualTo(1);
        }
    }
}
