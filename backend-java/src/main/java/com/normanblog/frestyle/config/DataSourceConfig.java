package com.normanblog.frestyle.config;

import com.zaxxer.hikari.HikariDataSource;
import javax.sql.DataSource;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.autoconfigure.condition.ConditionalOnExpression;
import org.springframework.boot.jdbc.DataSourceBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * 本番(Supabase)用の DataSource 設定。
 *
 * <p>環境変数 {@code DATABASE_URL} がセットされているとき**だけ** Postgres の DataSource を
 * 構築し、Spring Boot の H2 自動設定を上書きする。未設定(ローカル / テスト)では本 Bean は
 * 生成されず、application.properties の H2 がそのまま使われる。これにより本番では DATABASE_URL を
 * inject するだけで Postgres に切り替わる(既存 Go バックエンドと同じ運用)。
 */
@Configuration
public class DataSourceConfig {

  @Bean
  @ConditionalOnExpression("'${DATABASE_URL:}' != ''")
  DataSource dataSource(@Value("${DATABASE_URL}") String databaseUrl) {
    SupabaseUrl parsed = SupabaseUrl.parse(databaseUrl);
    HikariDataSource ds =
        DataSourceBuilder.create()
            .type(HikariDataSource.class)
            .driverClassName("org.postgresql.Driver")
            .url(parsed.jdbcUrl())
            .username(parsed.username())
            .password(parsed.password())
            .build();
    // Supabase Free の pooler は同時接続数が小さいため控えめに。
    ds.setMaximumPoolSize(5);
    ds.setPoolName("frestyle-supabase");
    return ds;
  }
}
