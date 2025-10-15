package com.example.FreStyle.controller;

import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;

import org.springframework.http.*;
import org.springframework.http.server.reactive.ServerHttpRequest;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/api")
public class UserInfoController {
  
  // Cookieの中に入っているid_tokenを取得して自分の情報を取得する
  @GetMapping("/me")
  public ResponseEntity<?> getUserinfo(ServerHttpRequest request) {
    
    
    // ストリームの処理になれていないのでコメントを残す
    
    // CookieからSESSIONを取得
    String Token = Optional.ofNullable(request.getCookies().get("SESSION")) // Optional.Nullableでnullでもエラーが出ないようにCookieを取得する
            .orElse(List.of()) // List<HttpCookie>の値がなかったらListのからの値を返す
            .stream() // コレクションクラス（今回のばあいはList）stream化する
            .findFirst() // 最初の値を得る
            .map(HttpCookie::getValue) // さらにそこから値（getValue()）を得る
            .orElse(null); // なかったらnullを返す
            
          
          if (Token == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
            .body(Map.of("error","未ログインです。"));
          }
          
          Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(Token);
          if (claimsOpt.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error","トークンが無効です。"));
          }
           
          JWTClaimsSet claims = claimsOpt.get();
          
          try {
            Map<String, String> userInfo = new HashMap<>();
            
            userInfo.put("name", claims.getStringClaim("name"));
            userInfo.put("email", claims.getStringClaim("email"));
            userInfo.put("sub", claims.getSubject());
            
            return ResponseEntity.ok(userInfo);
            
          } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "エラー情報の取得に失敗しました。"));
          }
  }
  
}
/* React side request
 * useEffect(() => {
  fetch('http://localhost:8080/api/me', {
    credentials: 'include',HttpOnlyクッキーの送信に必須なため
  })
    .then((res) => {
      if (!res.ok) {
        throw new Error('未ログイン');
      }
      return res.json();
    })
    .then((userInfo) => {
      console.log('ログイン済みユーザー:', userInfo);
    })
    .catch((err) => {
      console.log('ログインしていません');
      // 必要ならリダイレクト
    });
}, []);

 */