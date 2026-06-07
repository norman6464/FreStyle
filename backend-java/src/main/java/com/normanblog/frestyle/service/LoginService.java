package com.normanblog.frestyle.service;

import com.normanblog.frestyle.infra.cognito.CognitoTokenClient;
import com.normanblog.frestyle.infra.cognito.CognitoTokens;
import com.normanblog.frestyle.infra.cognito.IdTokenClaims;
import org.springframework.stereotype.Service;

/** ログインフロー(認可コード交換 / refresh)をオーケストレーションするサービス。 */
@Service
public class LoginService {

  private final CognitoTokenClient tokenClient;
  private final UserProvisioningService provisioning;

  public LoginService(CognitoTokenClient tokenClient, UserProvisioningService provisioning) {
    this.tokenClient = tokenClient;
    this.provisioning = provisioning;
  }

  /** 認可コードを token に交換し、users 行を upsert する。 */
  public LoginResult callback(String code, String invitationToken) {
    CognitoTokens tokens = tokenClient.exchangeAuthorizationCode(code);
    IdTokenClaims claims = IdTokenClaims.decode(tokens.idToken());
    boolean allowed = provisioning.upsertFromIdToken(claims, invitationToken);
    return new LoginResult(tokens, allowed);
  }

  /**
   * refresh_token で access_token を再発行し、id_token があれば role を同期する。
   *
   * <p>id_token があり upsert が false(招待が取り消された / ユーザー削除 等)なら allowed=false を返し、
   * controller 側で Cookie を破棄して 401 にする。id_token が無い場合は既存セッション前提で許可する。
   */
  public LoginResult refresh(String refreshToken) {
    CognitoTokens tokens = tokenClient.refreshAccessToken(refreshToken);
    boolean allowed = true;
    if (tokens.idToken() != null && !tokens.idToken().isBlank()) {
      allowed = provisioning.upsertFromIdToken(IdTokenClaims.decode(tokens.idToken()), null);
    }
    return new LoginResult(tokens, allowed);
  }

  /** callback の結果。allowed=false は招待ゲートで弾かれた新規ユーザー。 */
  public record LoginResult(CognitoTokens tokens, boolean allowed) {}
}
