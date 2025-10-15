package com.example.FreStyle.controller;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.http.server.reactive.ServerHttpResponse;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.reactive.function.BodyInserters;
import org.springframework.web.reactive.function.client.WebClient;

import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;

import reactor.core.publisher.Mono;

import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api/auth/cognito")
public class CognitoAuthController {

    @Value("${cognito.client-id}")
    private String clientId;

    @Value("${cognito.client-secret}")
    private String clientSecret;

    @Value("${cognito.redirect-uri}")
    private String redirectUri;

    @Value("${cognito.token-uri}")
    private String tokenUri;

    private final WebClient webClient;

    public CognitoAuthController(WebClient.Builder webClientBuilder) {
        this.webClient = webClientBuilder.build();
    }

    @PostMapping("/callback")
    public Mono<ResponseEntity<Map<String, String>>> callback(@RequestBody Map<String, String> body,
            ServerHttpResponse response) {
        String code = body.get("code");

        String basicAuthValue = Base64.getEncoder()
                .encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));

        MultiValueMap<String, String> formData = new LinkedMultiValueMap<>();
        formData.add("grant_type", "authorization_code");
        formData.add("code", code);
        formData.add("redirect_uri", redirectUri);
        formData.add("client_id", clientId);

        return webClient.post()
                .uri(tokenUri)
                .header(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_FORM_URLENCODED_VALUE)
                .header(HttpHeaders.AUTHORIZATION, "Basic " + basicAuthValue)
                .body(BodyInserters.fromFormData(formData))
                .retrieve() // 取得する(Cognitoが発行するアクセストークンidトークンが返ってくる)
                .bodyToMono(Map.class) // JSONをMAPで受け取る
                .flatMap(tokenResponse -> {
                    
                    String idToken = (String) tokenResponse.get("id_token");
                    
                    String accessToken = (String) tokenResponse.get("access_token");
                    // String refreshToken = (String) tokenResponse.get("refresh_token");

                    // IDトークンのクレームをパース
                    Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);

                    if (claimsOpt.isEmpty()) {
                        Map<String, String> errorResponse = new HashMap<>();
                        errorResponse.put("error", "IDトークンの解析に失敗しました。");

                        // 401へのステータスコードが返され、そのままreactが/loginへNavigateする
                        return Mono.just(ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                                .body(errorResponse));
                    }

                    JWTClaimsSet claims = claimsOpt.get();

                    Map<String, String> responseData = new HashMap<>();

                    // scopeで取得したidトークンの中身を分解
                    try {
                        String name = claims.getStringClaim("name");
                        String email = claims.getStringClaim("email");
                        String sub = claims.getSubject();
                        
                        System.out.println("ユーザー情報");
                        System.out.println("Name：" + name);
                        System.out.println("Email：" + email);
                        System.out.println("Sub：" + sub);
                        

                        responseData.put("name", name);
                        responseData.put("email", email);
                        responseData.put("sub", sub);
                        // accessTokenはデコードしない状態で保存する
                        responseData.put("accessToken", accessToken);

                        // 必要であればユーザーDBに登録 or セッション管理しようなど
                        // 今回はアクセストークンをHttpOnly Cookieに保存する
                    } catch (Exception e) {
                        e.printStackTrace();
                    }

                    return Mono.just(ResponseEntity.ok(responseData));

                })
                .onErrorResume(e -> {
                    Map<String, String> errorResponse = new HashMap<>();
                    errorResponse.put("error", "IDトークンの解析に失敗しました。");

                    return Mono.just(ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(errorResponse));
                });
    }
    
    
    @PostMapping("/logout")
    public ResponseEntity<Void> logout(ServerHttpResponse response) {
        ResponseCookie cookie = ResponseCookie.from("SESSION", "")
        .httpOnly(true)
        .secure(false) // 本番はtrueにする
        .sameSite("Lax")
        .maxAge(0) // すぐに期限切れ
        .path("/")
        .build();
        
        response.addCookie(cookie);
        
        return ResponseEntity.noContent().build();
    }
    
}
