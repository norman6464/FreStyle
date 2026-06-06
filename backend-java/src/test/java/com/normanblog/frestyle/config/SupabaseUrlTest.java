package com.normanblog.frestyle.config;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import org.junit.jupiter.api.Test;

/** libpq 形式 DATABASE_URL → JDBC 変換の検証。 */
class SupabaseUrlTest {

  @Test
  void parse_supabasePooler_disablesPreparedStatements() {
    SupabaseUrl r =
        SupabaseUrl.parse(
            "postgresql://postgres.abcdef:s3cret@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres?sslmode=require");

    assertThat(r.jdbcUrl())
        .isEqualTo(
            "jdbc:postgresql://aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres?sslmode=require&prepareThreshold=0");
    assertThat(r.username()).isEqualTo("postgres.abcdef");
    assertThat(r.password()).isEqualTo("s3cret");
  }

  @Test
  void parse_nonPooled_keepsPreparedStatements() {
    SupabaseUrl r = SupabaseUrl.parse("postgresql://user:pw@db.example.com:5432/postgres?sslmode=require");

    assertThat(r.jdbcUrl()).doesNotContain("prepareThreshold");
    assertThat(r.jdbcUrl()).isEqualTo("jdbc:postgresql://db.example.com:5432/postgres?sslmode=require");
  }

  @Test
  void parse_urlEncodedPassword_isDecoded() {
    // パスワードに記号が含まれ %-エンコードされている場合はデコードする。
    SupabaseUrl r = SupabaseUrl.parse("postgresql://user:p%40ss%2Fword@db.example.com:5432/postgres");
    assertThat(r.password()).isEqualTo("p@ss/word");
  }

  @Test
  void parse_missingCredentials_throws() {
    assertThatThrownBy(() -> SupabaseUrl.parse("postgresql://db.example.com:5432/postgres"))
        .isInstanceOf(IllegalArgumentException.class);
  }
}
