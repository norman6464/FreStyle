package com.normanblog.frestyle.cognito;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import java.util.ArrayList;
import java.util.Base64;
import java.util.List;

/**
 * id_token から必要なクレームを取り出した値。
 *
 * <p>id_token は Cognito の token endpoint から TLS 越しに受け取った直後のもので、署名検証は省く
 * (access_token は後続リクエストで JWKS 検証される)。ここでは payload を base64url デコードして
 * ユーザー作成に使う属性だけを読む(既存 Go 実装の DecodeClaims と同じ手法)。
 */
public record IdTokenClaims(String sub, String email, String name, List<String> groups) {

  private static final ObjectMapper MAPPER = new ObjectMapper();

  public static IdTokenClaims decode(String idToken) {
    String[] parts = idToken.split("\\.");
    if (parts.length < 2) {
      throw new IllegalArgumentException("invalid id_token");
    }
    try {
      JsonNode claims = MAPPER.readTree(base64UrlDecode(parts[1]));
      return new IdTokenClaims(
          text(claims, "sub"), text(claims, "email"), text(claims, "name"), groups(claims));
    } catch (RuntimeException | java.io.IOException e) {
      throw new IllegalArgumentException("invalid id_token", e);
    }
  }

  private static String text(JsonNode node, String field) {
    JsonNode v = node.get(field);
    return v == null || v.isNull() ? null : v.asText();
  }

  private static List<String> groups(JsonNode node) {
    JsonNode arr = node.get("cognito:groups");
    List<String> out = new ArrayList<>();
    if (arr != null && arr.isArray()) {
      arr.forEach(g -> out.add(g.asText()));
    }
    return out;
  }

  // JWT の base64url(パディング省略)を復元してデコードする。
  private static byte[] base64UrlDecode(String s) {
    int pad = (4 - s.length() % 4) % 4;
    return Base64.getUrlDecoder().decode(s + "=".repeat(pad));
  }
}
