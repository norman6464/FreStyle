package com.example.FreStyle.config;

import io.swagger.v3.oas.models.Components;
import io.swagger.v3.oas.models.OpenAPI;
import io.swagger.v3.oas.models.info.Contact;
import io.swagger.v3.oas.models.info.Info;
import io.swagger.v3.oas.models.info.License;
import io.swagger.v3.oas.models.security.SecurityRequirement;
import io.swagger.v3.oas.models.security.SecurityScheme;
import io.swagger.v3.oas.models.servers.Server;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.util.List;

/**
 * OpenAPI / Swagger UI のメタ情報定義。
 *
 * <p>Cookie ベース認証（{@code accessToken} HttpOnly Cookie）を SecurityScheme として登録し、
 * Swagger UI で「Try it out」する際に Cookie を自動付与する形を取る。</p>
 */
@Configuration
public class OpenApiConfig {

    private static final String COOKIE_AUTH = "cookieAuth";

    @Bean
    public OpenAPI freStyleOpenAPI() {
        return new OpenAPI()
                .info(new Info()
                        .title("FreStyle API")
                        .description("新卒IT エンジニア向けビジネスコミュニケーション練習アプリのバックエンド API。"
                                + "認証は AWS Cognito 発行の JWT を `accessToken` HttpOnly Cookie で送信する。")
                        .version("v1")
                        .contact(new Contact()
                                .name("FreStyle Team")
                                .url("https://normanblog.com"))
                        .license(new License()
                                .name("Proprietary")))
                .servers(List.of(
                        new Server().url("https://api.normanblog.com").description("Production"),
                        new Server().url("http://localhost:8080").description("Local")))
                .addSecurityItem(new SecurityRequirement().addList(COOKIE_AUTH))
                .components(new Components()
                        .addSecuritySchemes(COOKIE_AUTH, new SecurityScheme()
                                .type(SecurityScheme.Type.APIKEY)
                                .in(SecurityScheme.In.COOKIE)
                                .name("accessToken")
                                .description("Cognito 発行の JWT を HttpOnly Cookie として送信")));
    }
}
