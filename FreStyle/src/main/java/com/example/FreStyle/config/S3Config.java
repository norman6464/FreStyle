package com.example.FreStyle.config;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import software.amazon.awssdk.regions.Region;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;

/**
 * S3 Presigner 設定。
 *
 * <p>credentialsProvider を明示指定しないことで、AWS SDK の標準
 * クレデンシャルチェーンが以下の順に自動解決する:
 * <ol>
 *   <li>環境変数 AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY（ローカル開発・既存運用）</li>
 *   <li>~/.aws/credentials</li>
 *   <li>ECS Task Role（本番運用、推奨）</li>
 *   <li>EC2 Instance Profile</li>
 * </ol>
 * これにより、ローカル開発時は env vars、ECS 本番では Task Role と
 * 同一コードでシームレスに切り替わる。
 */
@Configuration
public class S3Config {

    @Value("${aws.region}")
    private String region;

    @Bean
    public S3Presigner s3Presigner() {
        return S3Presigner.builder()
                .region(Region.of(region))
                .build();
    }
}
