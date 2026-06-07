package com.normanblog.frestyle.infra.cognito;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.normanblog.frestyle.config.CognitoProperties;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.time.Duration;
import org.springframework.http.MediaType;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Component;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestClient;

/**
 * Cognito の OAuth2 token endpoint を叩いて、認可コード / refresh_token を
 * access/id/refresh token に交換する。
 *
 * <p>認証は body 方式(client_id + client_secret を form に入れる)に統一する。Basic ヘッダと
 * 併用すると invalid_client を返すことがあるため(既存 Go 実装と同じ)。
 */
@Component
public class CognitoTokenClient {

  private final CognitoProperties cognito;
  private final RestClient http;

  public CognitoTokenClient(CognitoProperties cognito) {
    this.cognito = cognito;
    SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
    factory.setConnectTimeout(Duration.ofSeconds(10));
    factory.setReadTimeout(Duration.ofSeconds(10));
    this.http = RestClient.builder().requestFactory(factory).build();
  }

  /** Hosted UI から戻った認可コードを token に交換する。 */
  public CognitoTokens exchangeAuthorizationCode(String code) {
    MultiValueMap<String, String> form = new LinkedMultiValueMap<>();
    form.add("grant_type", "authorization_code");
    form.add("code", code);
    form.add("redirect_uri", cognito.redirectUri());
    return exchange(form);
  }

  /** refresh_token を使って access_token を再発行する。 */
  public CognitoTokens refreshAccessToken(String refreshToken) {
    MultiValueMap<String, String> form = new LinkedMultiValueMap<>();
    form.add("grant_type", "refresh_token");
    form.add("refresh_token", refreshToken);
    return exchange(form);
  }

  private CognitoTokens exchange(MultiValueMap<String, String> form) {
    form.add("client_id", cognito.clientId());
    if (cognito.clientSecret() != null && !cognito.clientSecret().isBlank()) {
      form.add("client_secret", cognito.clientSecret());
    }

    return http.post()
        .uri(cognito.tokenUri())
        .contentType(MediaType.APPLICATION_FORM_URLENCODED)
        .body(form)
        .exchange(
            (request, response) -> {
              if (response.getStatusCode().isError()) {
                throw new TokenExchangeException(
                    response.getStatusCode().value(), readBody(response.getBody()));
              }
              TokenResponse body = response.bodyTo(TokenResponse.class);
              if (body == null) {
                throw new TokenExchangeException(response.getStatusCode().value(), "empty body");
              }
              return new CognitoTokens(
                  body.accessToken(), body.idToken(), body.refreshToken(), body.expiresIn());
            });
  }

  private static String readBody(java.io.InputStream in) {
    try {
      return new String(in.readAllBytes(), StandardCharsets.UTF_8);
    } catch (IOException e) {
      return "(unreadable body)";
    }
  }

  /** token endpoint の JSON(snake_case)レスポンス。 */
  private record TokenResponse(
      @JsonProperty("access_token") String accessToken,
      @JsonProperty("id_token") String idToken,
      @JsonProperty("refresh_token") String refreshToken,
      @JsonProperty("expires_in") int expiresIn) {}
}
