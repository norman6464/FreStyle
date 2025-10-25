package com.example.FreStyle.controller;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.*;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.reactive.function.BodyInserters;
import org.springframework.web.reactive.function.client.WebClient;

import com.example.FreStyle.form.ConfirmSignupForm;
import com.example.FreStyle.form.LoginForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;

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
    private final UserService userService;
    private final CognitoAuthService cognitoAuthService;

    public CognitoAuthController(WebClient.Builder webClientBuilder,UserService userService, CognitoAuthService cognitoAuthService) {
        this.webClient = webClientBuilder.build();
        this.userService = userService;
        this.cognitoAuthService = cognitoAuthService;
    }

    // Cognitoへサインアップ
    @PostMapping("/signup")
    public ResponseEntity<?> signup(@RequestBody SignupForm form) {
        try {
            // Cognitoに登録
            cognitoAuthService.signUpUser(form.getEmail(), form.getPassword(),
                    form.getName());
            // DBに保存
            userService.registerUser(form);

            return ResponseEntity.ok(Map.of("message", "サインアップ成功。確認メールを送信しました。"));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // メールで検証する
    @PostMapping("/confirm")
    public ResponseEntity<?> confirm(@RequestBody ConfirmSignupForm form) {
        try {
            // Cognitoで確認
            cognitoAuthService.confirmUserSignup(form.getEmail(), form.getCode());

            // DBユーザーを有効化
            userService.activeUser(form.getEmail());
            return ResponseEntity.ok(Map.of("message", "確認に成功しました。ログインできます。"));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // ログイン
    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody LoginForm form) {

        try {
            
            userService.checkUserIsActive(form.getEmail());
            Map<String, String> tokens = cognitoAuthService.login(form.getEmail(), form.getPassword());

            String idToken = tokens.get("idToken");
            
            String accessToken = tokens.get("accessToken");

            // デコードをしてクライアントに情報をわたす。
            Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);
            if (claimsOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なアクセスです。"));
            }

            JWTClaimsSet claims = claimsOpt.get();

            Map<String, String> responseData = new HashMap<>();
            
            try {
                
                // 取得したCognito id_tokenからsubjectを格納
                userService.registerCognitoSubject(claims.getSubject(), form.getEmail());
                responseData.put("email", claims.getStringClaim("email"));
                responseData.put("name", claims.getStringClaim("name"));
                responseData.put("sub", claims.getSubject());
                responseData.put("accessToken", accessToken);
                

            } catch (Exception e) {
                e.printStackTrace();
            }
            // アクセストークンのみを返却 or Cookieに保存も可能
            return ResponseEntity.ok(responseData);

        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // OIDCフロー
    @PostMapping("/callback")
    public ResponseEntity<Map<String, String>> callback(@RequestBody Map<String, String> body) {
        
        String code = body.get("code");

        String basicAuthValue = Base64.getEncoder()
                .encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));

        MultiValueMap<String, String> formData = new LinkedMultiValueMap<>();
        formData.add("grant_type", "authorization_code");
        formData.add("code", code);
        formData.add("redirect_uri", redirectUri);
        formData.add("client_id", clientId);

        Map<String, Object> tokenResponse = webClient.post()
                .uri(tokenUri)
                .header(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_FORM_URLENCODED_VALUE)
                .header(HttpHeaders.AUTHORIZATION, "Basic " + basicAuthValue)
                .body(BodyInserters.fromFormData(formData))
                .retrieve() // 取得する(Cognitoが発行するアクセストークンidトークンが返ってくる)
                .bodyToMono(new ParameterizedTypeReference<Map<String, Object>>() {
                }) // JSONをMAPで受け取る
                .block(); // Monoをブロックして同期処理（Spring MVCなので）

        if (tokenResponse == null) {
            Map<String, String> errorResponse = new HashMap<>();
            errorResponse.put("error", "トークン取得に失敗しました。");

            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(errorResponse);
        }

        String idToken = (String) tokenResponse.get("id_token");
        String accessToken = (String) tokenResponse.get("access_token");

        Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);

        if (claimsOpt.isEmpty()) {

            Map<String, String> errorResponse = new HashMap<>();
            errorResponse.put("error", "無効なリクエストです。");

            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(errorResponse);
        }

        JWTClaimsSet claims = claimsOpt.get();

        Map<String, String> responseData = new HashMap<>();
        
        try {
            
            String name = claims.getStringClaim("name");
            String email = claims.getStringClaim("email");
            String sub = claims.getSubject();
            
            System.out.println("email:" + email);
            System.out.println("sub:" + sub);
            userService.registerUserOIDC( "guest", email, sub);

            
            responseData.put("name", name);
            responseData.put("email", email);
            responseData.put("sub", sub);
            // responseData.put("idToken", idToken);
            responseData.put("accessToken", accessToken);
            
        } catch (Exception e) {
            e.printStackTrace();
        }

        return ResponseEntity.ok(responseData);
    }
}
