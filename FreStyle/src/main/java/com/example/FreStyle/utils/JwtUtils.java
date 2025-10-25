package com.example.FreStyle.utils;

import java.text.ParseException;
import java.util.Optional;

import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;

// Nimbus JWT + JOSE（ホゼ）ライブラリを使ってJWTをデコードして検証をするためのユーティリティーをつくる
// ClaimsはJWTのヘッダー、ペイロード、署名（Signature）の三つの中でペイロード部分に含まれているkey-value形式のデータの属性情報のこと
// これらのそれぞれをClaims（属性）
// {
//   "sub": "1234567890",       ← ユーザーID（Subject）
//   "name": "Taro Yamada",     ← ユーザー名
//   "admin": true,             ← 管理者フラグ
//   "iat": 1699999999          ← 発行時刻（Issued At）
// }

public class JwtUtils {
  
  public static Optional<JWTClaimsSet> decode(String token) {
    try {
      
      SignedJWT signedJWT = SignedJWT.parse(token);
      
      return Optional.of(signedJWT.getJWTClaimsSet());
      
    } catch(ParseException e) {
      return Optional.empty();
    }
  }
  
}
