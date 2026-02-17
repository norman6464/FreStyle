package com.example.FreStyle.controller;

import org.springframework.http.*;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.form.ConfirmSignupForm;
import com.example.FreStyle.form.ForgotPasswordForm;
import com.example.FreStyle.form.LoginForm;
import com.example.FreStyle.form.SignupForm;
import com.example.FreStyle.service.AuthCookieService;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.CognitoCallbackUseCase;
import com.example.FreStyle.usecase.CognitoConfirmUseCase;
import com.example.FreStyle.usecase.CognitoLoginUseCase;
import com.example.FreStyle.usecase.CognitoRefreshTokenUseCase;
import com.example.FreStyle.usecase.CognitoSignupUseCase;

import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import software.amazon.awssdk.services.cognitoidentityprovider.model.CodeMismatchException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.ExpiredCodeException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.InvalidPasswordException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.NotAuthorizedException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotConfirmedException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotFoundException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UsernameExistsException;

import java.util.Map;

import lombok.extern.slf4j.Slf4j;

@RestController
@RequestMapping("/api/auth/cognito")
@CrossOrigin(origins = "https://normanblog.com", allowCredentials = "true")
@RequiredArgsConstructor
@Slf4j
public class CognitoAuthController {

    private final CognitoAuthService cognitoAuthService;
    private final UserIdentityService userIdentityService;
    private final AuthCookieService authCookieService;
    private final CognitoLoginUseCase cognitoLoginUseCase;
    private final CognitoSignupUseCase cognitoSignupUseCase;
    private final CognitoConfirmUseCase cognitoConfirmUseCase;
    private final CognitoCallbackUseCase cognitoCallbackUseCase;
    private final CognitoRefreshTokenUseCase cognitoRefreshTokenUseCase;

    // -----------------------
    // サインアップ
    // -----------------------
    @PostMapping("/signup")
    public ResponseEntity<?> signup(@RequestBody SignupForm form) {
        log.info("POST /api/auth/cognito/signup - email: {}", form.getEmail());
        try {
            cognitoSignupUseCase.execute(form);
            return ResponseEntity.status(HttpStatus.CREATED)
                    .body(Map.of("message", "サインアップ成功。確認メールを送信しました。"));

        } catch (UsernameExistsException e) {
            return ResponseEntity.status(HttpStatus.CONFLICT)
                    .body(Map.of("error", "既にユーザーが存在しています。"));
        } catch (InvalidPasswordException e) {
            return ResponseEntity.badRequest().body(Map.of("error", "パスワードポリシーに違反しています。"));
        } catch (RuntimeException e) {
            log.error("/signup エラー: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // サインアップ確認
    // -----------------------
    @PostMapping("/confirm")
    public ResponseEntity<?> confirm(@RequestBody ConfirmSignupForm form) {
        log.info("POST /api/auth/cognito/confirm - email: {}", form.getEmail());
        try {
            cognitoConfirmUseCase.execute(form.getEmail(), form.getCode());
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
            log.error("/confirm エラー: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // ログイン
    // -----------------------
    @PostMapping("/login")
    public ResponseEntity<?> login(@RequestBody LoginForm form, HttpServletResponse response) {
        log.info("POST /api/auth/cognito/login - email: {}", form.getEmail());
        try {
            CognitoLoginUseCase.Result result = cognitoLoginUseCase.execute(form.getEmail(), form.getPassword());
            authCookieService.setAuthCookies(response,
                    result.accessToken(), result.refreshToken(), result.email(), result.cognitoUsername());
            return ResponseEntity.ok(Map.of("succes", "ログインできました。"));

        } catch (NotAuthorizedException e) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "メールアドレスまたはパスワードが間違っています。"));
        } catch (UserNotConfirmedException e) {
            return ResponseEntity.status(HttpStatus.FORBIDDEN)
                    .body(Map.of("error", "メール確認が完了していません。"));
        } catch (IllegalStateException e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "無効なアクセスです。"));
        } catch (RuntimeException e) {
            log.error("/login エラー: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // OIDCログイン Callback
    // -----------------------
    @PostMapping("/callback")
    public ResponseEntity<?> callback(@RequestBody Map<String, String> body, HttpServletResponse response) {
        log.info("POST /api/auth/cognito/callback");
        try {
            String code = body.get("code");
            CognitoCallbackUseCase.Result result = cognitoCallbackUseCase.execute(code);
            authCookieService.setAuthCookies(response,
                    result.accessToken(), result.refreshToken(), result.email(), result.cognitoUsername());
            return ResponseEntity.ok(Map.of("success", "ログインできました"));

        } catch (IllegalStateException e) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", e.getMessage()));
        } catch (Exception e) {
            log.error("/callback エラー: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", "server error: " + e.getMessage()));
        }
    }

    // -----------------------
    // ログアウト
    // -----------------------
    @PostMapping("/logout")
    public ResponseEntity<?> logout(@AuthenticationPrincipal Jwt jwt, HttpServletResponse response) {
        log.info("POST /api/auth/cognito/logout");
        String sub = jwt.getSubject();

        if (sub == null || sub.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "無効なリクエストです。"));
        }

        authCookieService.clearRefreshTokenCookie(response);
        return ResponseEntity.ok(Map.of("message", "ログアウトしました。"));
    }

    // -----------------------
    // パスワードリセット要求
    // -----------------------
    @PostMapping("/forgot-password")
    public ResponseEntity<?> forgotPassword(@Validated @RequestBody Map<String, String> body) {
        String email = body.get("email");
        log.info("POST /api/auth/cognito/forgot-password - email: {}", email);
        try {
            cognitoAuthService.forgotPassword(email);
            return ResponseEntity.ok(Map.of("message", "確認コードを送信しました。"));

        } catch (UserNotFoundException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "ユーザーが存在しません。"));
        } catch (RuntimeException e) {
            log.error("/forgot-password エラー: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // リフレッシュトークン
    // -----------------------
    @PostMapping("/refresh-token")
    public ResponseEntity<?> refreshToken(@CookieValue(name = "REFRESH_TOKEN", required = true) String refreshToken,
                                        @CookieValue(name = "EMAIL", required = true) String email,
                                        @CookieValue(name = "COGNITO_USERNAME", required = false) String cognitoUsername,
                                        HttpServletResponse response) {
        log.info("POST /api/auth/cognito/refresh-token");
        String username = (cognitoUsername != null && !cognitoUsername.isEmpty()) ? cognitoUsername : email;

        if (refreshToken == null || refreshToken.isEmpty()) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", "リフレッシュトークンが存在しません。"));
        }

        try {
            CognitoRefreshTokenUseCase.Result result = cognitoRefreshTokenUseCase.execute(refreshToken, username);
            authCookieService.setAuthCookies(response, result.newAccessToken(), refreshToken, email, username);
            return ResponseEntity.ok(Map.of("success", "更新完了"));

        } catch (RuntimeException e) {
            log.error("/refresh-token エラー: {}", e.getMessage(), e);
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED)
                    .body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // パスワードリセット確定
    // -----------------------
    @PostMapping("/confirm-forgot-password")
    public ResponseEntity<?> confirmForgotPassword(@Validated @RequestBody ForgotPasswordForm form) {
        log.info("POST /api/auth/cognito/confirm-forgot-password - email: {}", form.getEmail());
        try {
            cognitoAuthService.confirmForgotPassword(form.getEmail(), form.getCode(), form.getNewPassword());
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
            log.error("/confirm-forgot-password エラー: {}", e.getMessage(), e);
            return ResponseEntity.internalServerError()
                    .body(Map.of("error", e.getMessage()));
        }
    }

    // -----------------------
    // ユーザー情報取得
    // -----------------------
    @GetMapping("/me")
    public ResponseEntity<?> me(@AuthenticationPrincipal Jwt jwt) {
        String sub = jwt.getSubject();
        log.info("GET /api/auth/cognito/me - sub={}", sub);
        Integer id = userIdentityService.findUserBySub(sub).getId();
        return ResponseEntity.ok(Map.of("id", id));
    }
}
