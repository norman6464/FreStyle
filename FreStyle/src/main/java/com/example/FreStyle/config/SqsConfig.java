package com.example.FreStyle.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.scheduling.annotation.EnableScheduling;

import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.sqs.SqsClient;

/**
 * SQS Client 設定。
 *
 * <p>credentialsProvider を明示指定しないことで AWS SDK の標準
 * クレデンシャルチェーンが env vars → ~/.aws/credentials → ECS Task Role
 * → EC2 Instance Profile の順に自動解決する。S3Config と同方針。
 */
@Configuration
@EnableScheduling
public class SqsConfig {

    @Value("${aws.region}")
    private String region;

    @Bean
    public SqsClient sqsClient() {
        return SqsClient.builder()
                .region(Region.of(region))
                .build();
    }
}
