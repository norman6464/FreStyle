package com.example.FreStyle.config;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.amazonaws.xray.AWSXRay;
import com.amazonaws.xray.AWSXRayRecorderBuilder;

@DisplayName("XRayConfig テスト")
class XRayConfigTest {

    @AfterEach
    void tearDown() {
        AWSXRay.setGlobalRecorder(AWSXRayRecorderBuilder.defaultRecorder());
    }

    @Test
    @DisplayName("init実行後にグローバルレコーダーがnullでない")
    void initSetsGlobalRecorder() {
        XRayConfig config = new XRayConfig();
        config.init();

        assertThat(AWSXRay.getGlobalRecorder()).isNotNull();
    }
}
