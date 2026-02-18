package com.example.FreStyle.config;

import com.amazonaws.xray.AWSXRay;
import com.amazonaws.xray.AWSXRayRecorderBuilder;
import com.amazonaws.xray.plugins.ECSPlugin;
import com.amazonaws.xray.strategy.LogErrorContextMissingStrategy;
import com.amazonaws.xray.strategy.sampling.CentralizedSamplingStrategy;

import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;

import org.springframework.context.annotation.Configuration;

@Slf4j
@Configuration
public class XRayConfig {

    @PostConstruct
    public void init() {
        log.info("AWS X-Ray レコーダー初期化開始");

        AWSXRayRecorderBuilder builder = AWSXRayRecorderBuilder.standard()
                .withPlugin(new ECSPlugin())
                .withContextMissingStrategy(new LogErrorContextMissingStrategy())
                .withSamplingStrategy(new CentralizedSamplingStrategy());

        AWSXRay.setGlobalRecorder(builder.build());

        log.info("AWS X-Ray レコーダー初期化完了");
    }
}
