package com.example.FreStyle.controller;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.*;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.reactive.function.BodyInserters;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.security.oauth2.jwt.Jwt;
import com.example.FreStyle.form.ConfirmSignupForm;
import com.example.FreStyle.form.ForgotPasswordForm;
import com.example.FreStyle.form.LoginForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;

import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletResponse;

import java.nio.charset.StandardCharsets;
import java.text.ParseException;
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

    public CognitoAuthController(WebClient.Builder webClientBuilder, UserService userService,
            CognitoAuthService cognitoAuthService) {
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
    public ResponseEntity<?> login(@RequestBody LoginForm form, HttpServletResponse response) {

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
            userService.registerCognitoSubject(claims.getSubject(), form.getEmail());
            
            System.out.println("accessToken = " + accessToken);

            ResponseCookie cookie = ResponseCookie.from("ACCESS_TOKEN", accessToken)
                    .httpOnly(true)
                    .secure(true) // HTTPS 開発環境なら false
                    .path("/")
                    .maxAge(3600)
                    .sameSite("Lax") // クロスオリジンでも送信
                    .build();
            response.addHeader("Set-Cookie", cookie.toString());

            try {
                Map<String, Object> responseData = Map.of(
                        "email", claims.getStringClaim("email"),
                        "name", claims.getStringClaim("name"),
                        "sub", claims.getSubject(),
                        "message", "ログイン成功");
                        System.out.println("success");
                return ResponseEntity.ok(responseData);
            } catch (ParseException e) {
                return ResponseEntity.internalServerError()
                        .body(Map.of("error", "サーバーのエラーです。"));
            }

        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // OIDCフロー
    @PostMapping("/callback")
    public ResponseEntity<?> callback(@RequestBody Map<String, String> body, HttpServletResponse response) {

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

        try {

            String name = claims.getStringClaim("name");
            String email = claims.getStringClaim("email");
            String sub = claims.getSubject();

            userService.registerUserOIDC("guest", email, sub);

            ResponseCookie cookie = ResponseCookie.from("ACCESS_TOKEN", accessToken)
                    .httpOnly(true)
                    .secure(true) // HTTPS 開発環境なら false
                    .path("/")
                    .maxAge(3600)
                    .sameSite("Lax") // クロスオリジンでも送信
                    .build();
            response.addHeader("Set-Cookie", cookie.toString());

            Map<String, Object> responseData = Map.of(
                    "email", email,
                    "name", name,
                    "sub", sub);

            return ResponseEntity.ok(responseData);

        } catch (Exception e) {
            e.printStackTrace();
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", "サーバーのエラーが発生しました。"));
        }
    }

    @PostMapping("/logout")
    public ResponseEntity<?> logout(@AuthenticationPrincipal Jwt jwt, HttpServletResponse response) {
        String sub = jwt.getSubject();
        System.out.println("request GET /api/auth/cognito/logout");
        
        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なリクエストです。"));
        }
        
        Cookie cookie = new Cookie("ACCESS_TOKEN", null);
        cookie.setHttpOnly(true);
        cookie.setSecure(false);
        cookie.setPath("/");
        cookie.setMaxAge(60 * 60);
        return ResponseEntity.ok(Map.of("message", "ログアウトしました。"));
    }
    
    // パスワードリセットリクエスト
    @PostMapping("/forgot-password")
    public ResponseEntity<?> forgotPassword(@Validated @RequestBody Map<String, String> body) {
        
        String email = body.get("email");
        try {
            cognitoAuthService.forgotPassword(email);
            return ResponseEntity.ok(Map.of("message", "確認コードを送信しました。"));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }    
    }
    
    // パスワード更新
    @PostMapping("/confirm-forgot-password")
    public ResponseEntity<?> confirmForgotPassword(@Validated @RequestBody ForgotPasswordForm form) {
        String email = form.getEmail();
        String code = form.getCode();
        String newPassword = form.getNewPassword();
        
        try {
            cognitoAuthService.confirmForgotPassword(email, code, newPassword);
            return ResponseEntity.ok(Map.of("message", "パスワードをリセットしました。"));
        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }
    
}
