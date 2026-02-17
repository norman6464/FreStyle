package com.example.FreStyle.config;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.Timer;
import org.aspectj.lang.ProceedingJoinPoint;
import org.aspectj.lang.annotation.Around;
import org.aspectj.lang.annotation.Aspect;
import org.springframework.stereotype.Component;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@Slf4j
@Aspect
@Component
@RequiredArgsConstructor
public class RequestTimingAspect {

    private final MeterRegistry meterRegistry;

    @Around("execution(* com.example.FreStyle.controller..*(..))")
    public Object timeControllerMethods(ProceedingJoinPoint joinPoint) throws Throwable {
        return timeMethod(joinPoint, "controller");
    }

    @Around("execution(* com.example.FreStyle.service..*(..))")
    public Object timeServiceMethods(ProceedingJoinPoint joinPoint) throws Throwable {
        return timeMethod(joinPoint, "service");
    }

    @Around("execution(* com.example.FreStyle.usecase..*(..))")
    public Object timeUseCaseMethods(ProceedingJoinPoint joinPoint) throws Throwable {
        return timeMethod(joinPoint, "usecase");
    }

    private Object timeMethod(ProceedingJoinPoint joinPoint, String layer) throws Throwable {
        String className = joinPoint.getSignature().getDeclaringTypeName();
        String methodName = joinPoint.getSignature().getName();

        Timer.Sample sample = Timer.start(meterRegistry);
        try {
            Object result = joinPoint.proceed();
            log.info("[TIMING] {}.{} - {}ms", className, methodName,
                    sample.stop(buildTimer(className, methodName, layer)) / 1_000_000);
            return result;
        } catch (Throwable ex) {
            log.warn("[TIMING] {}.{} - {}ms (exception: {})", className, methodName,
                    sample.stop(buildTimer(className, methodName, layer)) / 1_000_000,
                    ex.getMessage());
            throw ex;
        }
    }

    private Timer buildTimer(String className, String methodName, String layer) {
        return Timer.builder("method.timed")
                .tag("class", className)
                .tag("method", methodName)
                .tag("layer", layer)
                .register(meterRegistry);
    }
}
