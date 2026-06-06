package com.normanblog.frestyle.config;

import java.net.URI;
import java.net.URLDecoder;
import java.nio.charset.StandardCharsets;

/**
 * Supabase / マネージド Postgres の libpq 形式 URL を JDBC 接続情報に変換する。
 *
 * <p>環境変数 {@code DATABASE_URL} は
 * {@code postgresql://user:password@host:6543/postgres?sslmode=require} という libpq 形式で
 * 渡ってくる(既存 Go バックエンドと同じ secret を共有)。JDBC ドライバはこの形式をそのまま
 * 受け取れないため、host/port/db と user/password に分解し JDBC URL を組み立てる。
 */
public record SupabaseUrl(String jdbcUrl, String username, String password) {

  /** libpq 形式の DATABASE_URL を解析する。host 名で pgbouncer を判定し prepared statement を無効化する。 */
  public static SupabaseUrl parse(String databaseUrl) {
    URI uri = URI.create(databaseUrl.trim());

    String host = uri.getHost();
    if (host == null) {
      throw new IllegalArgumentException("DATABASE_URL に host がありません");
    }
    int port = uri.getPort() == -1 ? 5432 : uri.getPort();
    String db = uri.getPath() == null || uri.getPath().isBlank() ? "/postgres" : uri.getPath();

    String userInfo = uri.getUserInfo();
    if (userInfo == null || !userInfo.contains(":")) {
      throw new IllegalArgumentException("DATABASE_URL に user:password がありません");
    }
    int sep = userInfo.indexOf(':');
    String username = decode(userInfo.substring(0, sep));
    String password = decode(userInfo.substring(sep + 1));

    // pgbouncer(Supabase Transaction pooler)はサーバ側 prepared statement を保持できないため、
    // host 名にプール識別子が含まれるときは prepareThreshold=0 で named statement を使わせない。
    boolean pooled = host.contains("pooler.supabase.com") || host.contains("pgbouncer");
    StringBuilder jdbc = new StringBuilder("jdbc:postgresql://").append(host).append(':').append(port).append(db);
    String query = uri.getQuery();
    jdbc.append('?').append(query == null || query.isBlank() ? "sslmode=require" : query);
    if (pooled) {
      jdbc.append("&prepareThreshold=0");
    }

    return new SupabaseUrl(jdbc.toString(), username, password);
  }

  private static String decode(String s) {
    return URLDecoder.decode(s, StandardCharsets.UTF_8);
  }
}
