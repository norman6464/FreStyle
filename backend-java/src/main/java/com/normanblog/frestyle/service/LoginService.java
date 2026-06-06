package com.normanblog.frestyle.service;

import com.normanblog.frestyle.cognito.CognitoTokenClient;
import com.normanblog.frestyle.cognito.CognitoTokens;
import com.normanblog.frestyle.cognito.IdTokenClaims;
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

  /** refresh_token で access_token を再発行し、id_token があれば role を同期する。 */
  public CognitoTokens refresh(String refreshToken) {
    CognitoTokens tokens = tokenClient.refreshAccessToken(refreshToken);
    if (tokens.idToken() != null && !tokens.idToken().isBlank()) {
      provisioning.upsertFromIdToken(IdTokenClaims.decode(tokens.idToken()), null);
    }
    return tokens;
  }

  /** callback の結果。allowed=false は招待ゲートで弾かれた新規ユーザー。 */
  public record LoginResult(CognitoTokens tokens, boolean allowed) {}
}
