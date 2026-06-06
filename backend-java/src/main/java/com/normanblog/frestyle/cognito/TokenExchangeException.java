package com.normanblog.frestyle.cognito;

/**
 * Cognito の token endpoint が non-2xx を返したときの例外。
 *
 * <p>{@code invalid_grant}(code 期限切れ / refresh 失効)や {@code redirect_uri_mismatch} など、
 * クライアント起因の失敗をこの型で表す。HTTP ステータスは呼び元(controller)で 401 等に変換する。
 */
public class TokenExchangeException extends RuntimeException {

  private final int httpStatus;

  public TokenExchangeException(int httpStatus, String body) {
    super("cognito token exchange failed: status=" + httpStatus + " body=" + body);
    this.httpStatus = httpStatus;
  }

  public int httpStatus() {
    return httpStatus;
  }
}
