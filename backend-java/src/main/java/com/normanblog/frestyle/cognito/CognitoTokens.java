package com.normanblog.frestyle.cognito;

/** Cognito の token endpoint レスポンス(必要項目のみ)。 */
public record CognitoTokens(
    String accessToken, String idToken, String refreshToken, int expiresIn) {}
