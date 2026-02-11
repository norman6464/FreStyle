package com.example.FreStyle.config;

import static org.assertj.core.api.Assertions.assertThat;

import io.micrometer.core.instrument.MeterRegistry;
import io.micrometer.core.instrument.simple.SimpleMeterRegistry;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;

class MetricsConfigTest {

    private MetricsConfig metricsConfig;

    @BeforeEach
    void setUp() {
        metricsConfig = new MetricsConfig();
    }

    @Nested
    @DisplayName("meterRegistryCustomizer - MeterRegistryのカスタマイズ")
    class MeterRegistryCustomizerTest {

        @Test
        @DisplayName("common tagsにappタグが設定される")
        void shouldSetAppTag() {
            MeterRegistry registry = new SimpleMeterRegistry();
            metricsConfig.meterRegistryCustomizer().customize(registry);

            registry.counter("test.counter").increment();

            assertThat(registry.get("test.counter").counter().getId().getTag("app"))
                    .isEqualTo("FreStyle");
        }

        @Test
        @DisplayName("common tagsにenvタグが設定される")
        void shouldSetEnvTag() {
            MeterRegistry registry = new SimpleMeterRegistry();
            metricsConfig.meterRegistryCustomizer().customize(registry);

            registry.counter("test.counter").increment();

            assertThat(registry.get("test.counter").counter().getId().getTag("env"))
                    .isEqualTo("production");
        }

        @Test
        @DisplayName("カスタマイザーがnullを返さない")
        void shouldReturnNonNullCustomizer() {
            assertThat(metricsConfig.meterRegistryCustomizer()).isNotNull();
        }
    }
}
