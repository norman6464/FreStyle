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

import com.example.FreStyle.entity.AccessToken;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ConfirmSignupForm;
import com.example.FreStyle.form.ForgotPasswordForm;
import com.example.FreStyle.form.LoginForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;

import jakarta.servlet.http.HttpServletResponse;
import software.amazon.awssdk.services.cognitoidentityprovider.model.CodeMismatchException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.ExpiredCodeException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.InvalidPasswordException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.NotAuthorizedException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotConfirmedException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotFoundException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UsernameExistsException;

import java.nio.charset.StandardCharsets;
import java.text.ParseException;
import java.util.Base64;
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
    private final UserIdentityService userIdentityService;
    private final AccessTokenService accessTokenService;

    public CognitoAuthController(WebClient.Builder webClientBuilder,
        UserService userService,
        CognitoAuthService cognitoAuthService,
        UserIdentityService userIdentityService,
        AccessTokenService accessTokenService) {
        this.webClient = webClientBuilder.build();
        this.userService = userService;
        this.cognitoAuthService = cognitoAuthService;
        this.userIdentityService = userIdentityService;
        this.accessTokenService = accessTokenService;
    }

    // -----------------------
    // サインアップ
    // -----------------------
    @PostMapping("/signup")
    public ResponseEntity<?> signup(@RequestBody SignupForm form) {
        try {
            cognitoAuthService.signUpUser(form.getEmail(), form.getPassword(), form.getName());
            userService.registerUser(form);

            return ResponseEntity.status(HttpStatus.CREATED)
                    .body(Map.of("message", "サインアップ成功。確認メールを送信しました。"));

        } catch (UsernameExistsException e) {
            return ResponseEntity.status(HttpStatus.CONFLICT)
                    .body(Map.of("error", "既にユーザーが存在しています。"));

        } catch (InvalidPasswordException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "パスワードポリシーに違反しています。"));

        } catch (RuntimeException e) {
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // サインアップ確認
    // -----------------------
    @PostMapping("/confirm")
    public ResponseEntity<?> confirm(@RequestBody ConfirmSignupForm form) {
        try {
            cognitoAuthService.confirmUserSignup(form.getEmail(), form.getCode());
            userService.activeUser(form.getEmail());

            return ResponseEntity.ok(Map.of("message", "確認に成功しました。ログインできます。"));

        } catch (CodeMismatchException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "確認コードが正しくありません。"));

        } catch (ExpiredCodeException e) {
            return ResponseEntity.status(HttpStatus.GONE)
                    .body(Map.of("error", "確認コードの有効期限が切れています。"));

        } catch (UserNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "ユーザーが存在しません。"));

        } catch (RuntimeException e) {
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // ログイン
    // -----------------------
    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody LoginForm form, HttpServletResponse response) {

        try {
            userService.checkUserIsActive(form.getEmail());
            Map<String, String> tokens = cognitoAuthService.login(form.getEmail(), form.getPassword());

            String idToken = tokens.get("idToken");
            String accessToken = tokens.get("accessToken");
            String refreshToken = tokens.get("refreshToken");

            Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);
            if (claimsOpt.isEmpty()) {
                return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なアクセスです。"));
            }

            JWTClaimsSet claims = claimsOpt.get();

            User user = userService.findUserByEmail(form.getEmail());
            userIdentityService.registerUserIdentity(user, claims.getIssuer(), claims.getSubject());
            
            setAuthCookies(response, accessToken, refreshToken);

            accessTokenService.saveTokens(user, accessToken, refreshToken);

            return ResponseEntity.ok(Map.of("succes", "ログインできました。"));

        } catch (RuntimeException e) {
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // OIDCログイン Callback
    // -----------------------
    @PostMapping("/callback")
    public ResponseEntity<?> callback(@RequestBody Map<String, String> body, HttpServletResponse response) {
        System.out.println("[CognitoAuthController /callback] Callback endpoint called");
        String code = body.get("code");
        System.out.println("[CognitoAuthController /callback] Authorization code received: " + 
                          (code != null ? code.substring(0, Math.min(20, code.length())) + "..." : "null"));

        String basicAuthValue = Base64.getEncoder()
                .encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));

        MultiValueMap<String, String> formData = new LinkedMultiValueMap<>();
        formData.add("grant_type", "authorization_code");
        formData.add("code", code);
        formData.add("redirect_uri", redirectUri);
        formData.add("client_id", clientId);

        System.out.println("[CognitoAuthController /callback] Requesting token from Cognito");
        Map<String, Object> tokenResponse = webClient.post()
                .uri(tokenUri)
                .header(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_FORM_URLENCODED_VALUE)
                .header(HttpHeaders.AUTHORIZATION, "Basic " + basicAuthValue)
                .body(BodyInserters.fromFormData(formData))
                .retrieve()
                .bodyToMono(new ParameterizedTypeReference<Map<String, Object>>() {
                })
                .block();

        if (tokenResponse == null) {
            System.err.println("[CognitoAuthController /callback] ERROR: tokenResponse is null");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "トークン取得に失敗しました。"));
        }

        System.out.println("[CognitoAuthController /callback] トークン取得成功");

        String idToken = (String) tokenResponse.get("id_token");
        String accessToken = (String) tokenResponse.get("access_token");
        String refreshToken = (String) tokenResponse.get("refresh_token");

        System.out.println("[CognitoAuthController /callback] Token types - accessToken: " + 
                          (accessToken != null ? "✓" : "null") + 
                          ", refreshToken: " + (refreshToken != null ? "✓" : "null"));

        Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);
        if (claimsOpt.isEmpty()) {
            System.err.println("[CognitoAuthController /callback] ERROR: Failed to decode idToken");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "無効なリクエストです。"));
        }

        try {
            JWTClaimsSet claims = claimsOpt.get();
            String email = claims.getStringClaim("email");
            String sub = claims.getSubject();

            System.out.println("[CognitoAuthController /callback] User info - email: " + email + ", sub: " + sub);

            boolean isGoogle = claims.getClaim("identities") != null;
            String provider = isGoogle ? "google" : "cognito";

            System.out.println("[CognitoAuthController /callback] Registering user - provider: " + provider);
            userService.registerUserOIDC("guest", email, provider, sub);

            // httpOnlyCookieの設定
            System.out.println("[CognitoAuthController /callback] Setting auth cookies");
            setAuthCookies(response, accessToken, refreshToken);

            return ResponseEntity.ok(Map .of("success","ログインできました"));

        } catch (Exception e) {
            System.out.println("[CognitoAuthController /callback] ERROR: " + e.getClass().getSimpleName() + " - " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", "server error: " + e.getMessage()));
        }
    }

    // -----------------------
    // ログアウト
    // -----------------------
    @PostMapping("/logout")
    public ResponseEntity<?> logout(@AuthenticationPrincipal Jwt jwt, HttpServletResponse response) {
        String sub = jwt.getSubject();

        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "無効なリクエストです。"));
        }

        ResponseCookie cookie = ResponseCookie.from("REFRESH_TOKEN", null)
                .httpOnly(true)
                .secure(false)
                .path("/")
                .maxAge(0)
                .sameSite("None")
                .build();
        response.addHeader("Set-Cookie", cookie.toString());

        return ResponseEntity.ok(Map.of("message", "ログアウトしました。"));
    }

    // -----------------------
    // パスワードリセット要求
    // -----------------------
    @PostMapping("/forgot-password")
    public ResponseEntity<?> forgotPassword(@Validated @RequestBody Map<String, String> body) {

        String email = body.get("email");

        try {
            cognitoAuthService.forgotPassword(email);
            return ResponseEntity.ok(Map.of("message", "確認コードを送信しました。"));

        } catch (UserNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "ユーザーが存在しません。"));

        } catch (RuntimeException e) {
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // リフレッシュトークンを使用をしてアクセストークン、IDトークンの再発行を行う
    // -----------------------
    @PostMapping("/refresh-token")
    public ResponseEntity<?> refreshToken(@CookieValue(name = "REFRESH_TOKEN", required = true) String refreshToken,
                    HttpServletResponse response) {

        System.out.println("[CognitoAuthController /refresh-token] Endpoint called");
        System.out.println("[CognitoAuthController /refresh-token] REFRESH_TOKEN cookie: " + 
                          (refreshToken != null ? refreshToken.substring(0, Math.min(20, refreshToken.length())) + "..." : "null"));

        if (refreshToken == null || refreshToken.isEmpty()) {
            System.out.println("[CognitoAuthController /refresh-token] ERROR: REFRESH_TOKEN cookie is null or empty");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "リフレッシュトークンが存在しません。"));
        }

        try {

            AccessToken accessTokenEntity = accessTokenService.findAccessTokenByRefreshToken(refreshToken);

            System.out.println("[CognitoAuthController /refresh-token] Attempting to refresh access token");
            Map<String, String> tokens = cognitoAuthService.refreshAccessToken(refreshToken);
            System.out.println("[CognitoAuthController /refresh-token] Successfully refreshed tokens");

            accessTokenService.updateTokens(
                    accessTokenEntity,
                    tokens.get("accessToken")
            );
            
            setAuthCookies(response, tokens.get("accessToken"), refreshToken);
            return ResponseEntity.ok(Map.of("success","更新完了"));

        } catch (RuntimeException e) {
            System.out.println("[CognitoAuthController /refresh-token] ERROR: " + e.getClass().getSimpleName() + " - " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // パスワードリセット確定
    // -----------------------
    @PostMapping("/confirm-forgot-password")
    public ResponseEntity<?> confirmForgotPassword(@Validated @RequestBody ForgotPasswordForm form) {
        String email = form.getEmail();
        String code = form.getCode();
        String newPassword = form.getNewPassword();

        try {
            cognitoAuthService.confirmForgotPassword(email, code, newPassword);

            return ResponseEntity.ok(Map.of("message", "パスワードをリセットしました。"));

        } catch (UserNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "ユーザーが存在しません。"));

        } catch (CodeMismatchException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "確認コードが正しくありません。"));

        } catch (ExpiredCodeException e) {
            return ResponseEntity.status(HttpStatus.GONE)
                    .body(Map.of("error", "確認コードの有効期限が切れています。"));

        } catch (InvalidPasswordException e) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "パスワードポリシーに違反しています。"));

        } catch (RuntimeException e) {
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", e.getMessage()));
        }
    }


    
    // -----------------------
    // Cookie格納メソッド
    // -----------------------
    @GetMapping("/me")
    public ResponseEntity<?> me(@AuthenticationPrincipal Jwt jwt) {
        System.out.println("[CognitoAuthController /me] Endpoint called");
        
        if (jwt == null) {
            System.out.println("[CognitoAuthController /me] ERROR: JWT is null - user not authenticated");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "認証されていません"));
        }
        
        System.out.println("[CognitoAuthController /me] JWT Principal: " + jwt.toString());
        
        String sub = jwt.getSubject();
        System.out.println("[CognitoAuthController /me] JWT Subject (sub): " + sub);

        if (sub == null || sub.isEmpty()) {
            System.out.println("[CognitoAuthController /me] ERROR: JWT subject is null or empty");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "無効なリクエストです。"));
        }

        try {
            System.out.println("[CognitoAuthController /me] Finding user with sub: " + sub);
            Integer id = userIdentityService.findUserBySub(sub).getId();
            System.out.println("[CognitoAuthController /me] User found: " + id);

            return ResponseEntity.ok(Map.of("id",id));

        } catch (RuntimeException e) {
            System.out.println("[CognitoAuthController /me] ERROR: " + e.getClass().getSimpleName() + " - " + e.getMessage());
            e.printStackTrace();
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", e.getMessage()));
        }
    } 


    // -----------------------
    // Cookie格納メソッド
    // -----------------------
    private void setAuthCookies(
        HttpServletResponse response,
        String accessToken,
        String refreshToken
    ) {
    ResponseCookie accessCookie = ResponseCookie.from("ACCESS_TOKEN", accessToken)
            .httpOnly(true)
            .secure(false) // 開発環境: false、本番環境: true
            .path("/")
            .maxAge(60 * 15) // 15分
            .sameSite("Lax") // 開発環境: Lax、本番環境: None
            .build();

    ResponseCookie refreshCookie = ResponseCookie.from("REFRESH_TOKEN", refreshToken)
            .httpOnly(true)
            .secure(false) // 開発環境: false、本番環境: true
            .path("/")
            .maxAge(60 * 60 * 24 * 7) // 7日
            .sameSite("Lax") // 開発環境: Lax、本番環境: None
            .build();

    System.out.println("[setAuthCookies] Setting cookies - ACCESS_TOKEN and REFRESH_TOKEN");
    response.addHeader(HttpHeaders.SET_COOKIE, accessCookie.toString());
    response.addHeader(HttpHeaders.SET_COOKIE, refreshCookie.toString());
}


    


}
