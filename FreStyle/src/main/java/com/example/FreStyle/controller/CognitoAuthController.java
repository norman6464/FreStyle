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
import com.example.FreStyle.service.CognitoLoginService;
import com.example.FreStyle.service.CognitoSignupService;
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
    private final CognitoSignupService signupService;
    private final CognitoLoginService loginService;
    private final UserService userService;

    public CognitoAuthController(WebClient.Builder webClientBuilder, CognitoSignupService signupService,
            CognitoLoginService loginService, UserService userService) {
        this.webClient = webClientBuilder.build();
        this.signupService = signupService;
        this.loginService = loginService;
        this.userService = userService;
    }

    // Cognitoへサインアップ
    @PostMapping("/signup")
    public ResponseEntity<?> signup(@RequestBody SignupForm form) {
        System.out.println(form.getName());
        System.out.println(form.getEmail());
        System.out.println(form.getPassword());

        try {
            // Cognitoに登録
            signupService.signUpUser(form.getEmail(), form.getPassword(),
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
            signupService.confirmUserSignup(form.getEmail(), form.getCode());

            // DBユーザーを有効化
            userService.activeUserByEmail(form.getEmail());
            return ResponseEntity.ok(Map.of("message", "確認に成功しました。ログインできます。"));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // ログイン
    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody LoginForm form) {

        // System.out.println(form.getEmail());
        // System.out.println(form.getPassword());

        try {
            Map<String, String> tokens = loginService.login(form.getEmail(), form.getPassword());

            String id_token = tokens.get("id_token");
            
            System.out.println("id_token value " + id_token);
            String access_token = tokens.get("access_token");

            // デコードをしてクライアントに情報をわたす。
            Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(id_token);
            if (claimsOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "failed id_token parse"));
            }

            JWTClaimsSet claims = claimsOpt.get();

            Map<String, String> responseData = new HashMap<>();

            try {
                
                responseData.put("email", claims.getStringClaim("email"));
                responseData.put("name", claims.getStringClaim("name"));
                responseData.put("sub", claims.getSubject());
                responseData.put("access_token", access_token);
                // このメソッドをCognitoLoginServiceで記述するか模索中
                // userService.checkUserIsActiveByEmail(form.getEmail());

            } catch (Exception e) {
                e.printStackTrace();
            }
            // access_tokenのみを返却 or Cookieに保存も可能
            return ResponseEntity.ok(responseData);

        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // OIDCフロー
    @PostMapping("/callback")
    public ResponseEntity<Map<String, String>> callback(@RequestBody Map<String, String> body) {

        System.out.println(body);
        // System.out.println("リクエストを受け付けました。");

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
        // String refreshToken = (String) tokenResponse.get("refresh_token");

        // IDトークンのクレームをパース
        Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);

        if (claimsOpt.isEmpty()) {

            Map<String, String> errorResponse = new HashMap<>();
            errorResponse.put("error", "failed id_token parse");

            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(errorResponse);
        }

        JWTClaimsSet claims = claimsOpt.get();

        Map<String, String> responseData = new HashMap<>();

        // scopeで取得したidトークンの中身を分解
        try {
            String name = claims.getStringClaim("name");
            String email = claims.getStringClaim("email");
            String sub = claims.getSubject();

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

        return ResponseEntity.ok(responseData);
    }
}
