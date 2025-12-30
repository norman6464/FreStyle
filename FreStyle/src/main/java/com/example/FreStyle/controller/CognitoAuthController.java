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

import com.example.FreStyle.entity.User;
import com.example.FreStyle.form.ConfirmSignupForm;
import com.example.FreStyle.form.ForgotPasswordForm;
import com.example.FreStyle.form.LoginForm;
import com.example.FreStyle.form.SignupForm;
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

    public CognitoAuthController(WebClient.Builder webClientBuilder, UserService userService,
            CognitoAuthService cognitoAuthService, UserIdentityService userIdentityService) {
        this.webClient = webClientBuilder.build();
        this.userService = userService;
        this.cognitoAuthService = cognitoAuthService;
        this.userIdentityService = userIdentityService;
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

            ResponseCookie cookie = ResponseCookie.from("REFRESH_TOKEN", refreshToken)
                    .httpOnly(true).secure(false).path("/").maxAge(3600).sameSite("None").build();
            response.addHeader("Set-Cookie", cookie.toString());

            Map<String, Object> responseData = Map.of(
                    "email", claims.getStringClaim("email"),
                    "name", claims.getStringClaim("name"),
                    "sub", claims.getSubject(),
                    "accessToken", accessToken);

            return ResponseEntity.ok(responseData);

        } catch (UserNotFoundException e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "ユーザーが存在しません。"));

        } catch (NotAuthorizedException e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "メールアドレスまたはパスワードが違います。"));

        } catch (UserNotConfirmedException e) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "メールアドレスが確認されていません。"));

        } catch (ParseException e) {
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", "トークンの解析中にエラーが発生しました。"));

        } catch (RuntimeException e) {
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // OIDCログイン Callback
    // -----------------------
    @PostMapping("/callback")
    public ResponseEntity<?> callback(@RequestBody Map<String, String> body, HttpServletResponse response) {
        System.out.println("/api/auth/cognito/callback");
        System.out.println("start callback to oidc");
        String code = body.get("code");
        System.out.println("認可コード: " + code);

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
                .retrieve()
                .bodyToMono(new ParameterizedTypeReference<Map<String, Object>>() {
                })
                .block();

        if (tokenResponse == null) {
            System.err.println("Failed take for token");
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "トークン取得に失敗しました。"));
        }

        System.out.println("トークン取得成功");

        String idToken = (String) tokenResponse.get("id_token");
        String accessToken = (String) tokenResponse.get("access_token");
        String refreshToken = (String) tokenResponse.get("refresh_token");

        Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);
        if (claimsOpt.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "無効なリクエストです。"));
        }

        try {
            JWTClaimsSet claims = claimsOpt.get();
            String name = claims.getStringClaim("name");
            String email = claims.getStringClaim("email");
            String sub = claims.getSubject();

            boolean isGoogle = claims.getClaim("identities") != null;
            String provider = isGoogle ? "google" : "cognito";

            userService.registerUserOIDC("guest", email, provider, sub);

            // 一時的にローカルで開発をしているためsecureをfalseにしている
            ResponseCookie cookie = ResponseCookie.from("REFRESH_TOKEN", refreshToken)
                    .httpOnly(true)
                    .secure(false
                    )
                    .path("/")
                    .maxAge(3600)
                    .sameSite("None")
                    .build();
            response.addHeader("Set-Cookie", cookie.toString());

            Map<String, Object> responseData = Map.of(
                    "email", email,
                    "name", name,
                    "sub", sub,
                    "accessToken", accessToken);

            return ResponseEntity.ok(responseData);

        } catch (Exception e) {
            System.out.println("oidc callback error " + e.getMessage());
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
    // トークンリフレッシュ
    // -----------------------
    @PostMapping("/refresh-token")
    public ResponseEntity<?> refreshToken(@CookieValue(name = "REFRESH_TOKEN", required = false) String refreshToken) {

        if (refreshToken == null || refreshToken.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "リフレッシュトークンが存在しません。"));
        }

        try {
            Map<String, String> tokens = cognitoAuthService.refreshAccessToken(refreshToken);

            return ResponseEntity.ok(Map.of("accessToken", tokens.get("accessToken")));

        } catch (RuntimeException e) {
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
}
