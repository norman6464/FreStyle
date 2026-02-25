package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.entity.User;
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
import software.amazon.awssdk.services.cognitoidentityprovider.model.CodeMismatchException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.ExpiredCodeException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.InvalidPasswordException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UserNotFoundException;
import software.amazon.awssdk.services.cognitoidentityprovider.model.UsernameExistsException;

@ExtendWith(MockitoExtension.class)
@DisplayName("CognitoAuthController")
class CognitoAuthControllerTest {

    @Mock private CognitoAuthService cognitoAuthService;
    @Mock private UserIdentityService userIdentityService;
    @Mock private AuthCookieService authCookieService;
    @Mock private CognitoLoginUseCase cognitoLoginUseCase;
    @Mock private CognitoSignupUseCase cognitoSignupUseCase;
    @Mock private CognitoConfirmUseCase cognitoConfirmUseCase;
    @Mock private CognitoCallbackUseCase cognitoCallbackUseCase;
    @Mock private CognitoRefreshTokenUseCase cognitoRefreshTokenUseCase;
    @Mock private HttpServletResponse httpResponse;

    private CognitoAuthController controller;

    @BeforeEach
    void setUp() {
        controller = new CognitoAuthController(
            cognitoAuthService, userIdentityService, authCookieService,
            cognitoLoginUseCase, cognitoSignupUseCase, cognitoConfirmUseCase,
            cognitoCallbackUseCase, cognitoRefreshTokenUseCase
        );
    }

    private Jwt mockJwt(String sub) {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn(sub);
        return jwt;
    }

    @Nested
    @DisplayName("signup")
    class Signup {

        @Test
        @DisplayName("サインアップ成功で201を返す")
        void returnsCreatedOnSuccess() {
            SignupForm form = new SignupForm("test@example.com", "password123", "テスト");

            ResponseEntity<?> response = controller.signup(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CREATED);
            verify(cognitoSignupUseCase).execute(form);
        }

        @Test
        @DisplayName("UsernameExistsExceptionで409を返す")
        void returnsConflictOnUsernameExists() {
            SignupForm form = new SignupForm("test@example.com", "password123", "テスト");
            doThrow(UsernameExistsException.builder().message("exists").build())
                .when(cognitoSignupUseCase).execute(any(SignupForm.class));

            ResponseEntity<?> response = controller.signup(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.CONFLICT);
        }

        @Test
        @DisplayName("InvalidPasswordExceptionで400を返す")
        void returnsBadRequestOnInvalidPassword() {
            SignupForm form = new SignupForm("test@example.com", "weak", "テスト");
            doThrow(InvalidPasswordException.builder().message("invalid").build())
                .when(cognitoSignupUseCase).execute(any(SignupForm.class));

            ResponseEntity<?> response = controller.signup(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }

        @Test
        @DisplayName("RuntimeExceptionで500を返し内部エラー情報を漏洩しない")
        @SuppressWarnings("unchecked")
        void returnsInternalServerErrorOnRuntimeException() {
            SignupForm form = new SignupForm("test@example.com", "password123", "テスト");
            doThrow(new RuntimeException("Internal DB connection pool exhausted"))
                .when(cognitoSignupUseCase).execute(any(SignupForm.class));

            ResponseEntity<?> response = controller.signup(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.INTERNAL_SERVER_ERROR);
            Map<String, String> body = (Map<String, String>) response.getBody();
            assertThat(body.get("error")).doesNotContain("DB connection pool");
        }
    }

    @Nested
    @DisplayName("confirm")
    class Confirm {

        @Test
        @DisplayName("確認成功で200を返す")
        void returnsOkOnSuccess() {
            ConfirmSignupForm form = new ConfirmSignupForm("test@example.com", "123456");

            ResponseEntity<?> response = controller.confirm(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(cognitoConfirmUseCase).execute("test@example.com", "123456");
        }

        @Test
        @DisplayName("CodeMismatchExceptionで400を返す")
        void returnsBadRequestOnCodeMismatch() {
            ConfirmSignupForm form = new ConfirmSignupForm("test@example.com", "wrong");
            doThrow(CodeMismatchException.builder().message("mismatch").build())
                .when(cognitoConfirmUseCase).execute(anyString(), anyString());

            ResponseEntity<?> response = controller.confirm(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }

        @Test
        @DisplayName("ExpiredCodeExceptionで410を返す")
        void returnsGoneOnExpiredCode() {
            ConfirmSignupForm form = new ConfirmSignupForm("test@example.com", "expired");
            doThrow(ExpiredCodeException.builder().message("expired").build())
                .when(cognitoConfirmUseCase).execute(anyString(), anyString());

            ResponseEntity<?> response = controller.confirm(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.GONE);
        }

        @Test
        @DisplayName("UserNotFoundExceptionで404を返す")
        void returnsNotFoundOnUserNotFound() {
            ConfirmSignupForm form = new ConfirmSignupForm("unknown@example.com", "123456");
            doThrow(UserNotFoundException.builder().message("not found").build())
                .when(cognitoConfirmUseCase).execute(anyString(), anyString());

            ResponseEntity<?> response = controller.confirm(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
        }
    }

    @Nested
    @DisplayName("callback")
    class Callback {

        @Test
        @DisplayName("コールバック成功で200を返す")
        void returnsOkOnSuccess() {
            CognitoCallbackUseCase.Result result = new CognitoCallbackUseCase.Result(
                    "access-token", "refresh-token", "test@example.com", "cognito-user");
            when(cognitoCallbackUseCase.execute("auth-code")).thenReturn(result);

            ResponseEntity<?> response = controller.callback(Map.of("code", "auth-code"), httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(authCookieService).setAuthCookies(httpResponse,
                    "access-token", "refresh-token", "test@example.com", "cognito-user");
        }

        @Test
        @DisplayName("IllegalStateExceptionで401を返す")
        void returnsUnauthorizedOnIllegalState() {
            when(cognitoCallbackUseCase.execute(anyString()))
                    .thenThrow(new IllegalStateException("トークン取得に失敗しました。"));

            ResponseEntity<?> response = controller.callback(Map.of("code", "bad-code"), httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }

    @Nested
    @DisplayName("refreshToken")
    class RefreshToken {

        @Test
        @DisplayName("トークン更新成功で200を返す")
        void returnsOkOnSuccess() {
            CognitoRefreshTokenUseCase.Result result = new CognitoRefreshTokenUseCase.Result("new-access-token");
            when(cognitoRefreshTokenUseCase.execute("refresh-token", "testuser")).thenReturn(result);

            ResponseEntity<?> response = controller.refreshToken(
                    "refresh-token", "test@example.com", "testuser", httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(authCookieService).setAuthCookies(httpResponse,
                    "new-access-token", "refresh-token", "test@example.com", "testuser");
        }

        @Test
        @DisplayName("空のリフレッシュトークンで401を返す")
        void returnsUnauthorizedOnEmptyRefreshToken() {
            ResponseEntity<?> response = controller.refreshToken(
                    "", "test@example.com", "testuser", httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("RuntimeExceptionで401を返す")
        void returnsUnauthorizedOnRuntimeException() {
            when(cognitoRefreshTokenUseCase.execute(anyString(), anyString()))
                    .thenThrow(new RuntimeException("無効なリフレッシュトークンです。"));

            ResponseEntity<?> response = controller.refreshToken(
                    "invalid-token", "test@example.com", "testuser", httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }

    @Nested
    @DisplayName("forgotPassword")
    class ForgotPassword {

        @Test
        @DisplayName("パスワードリセット要求成功で200を返す")
        void returnsOkOnSuccess() {
            ResponseEntity<?> response = controller.forgotPassword(Map.of("email", "test@example.com"));

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(cognitoAuthService).forgotPassword("test@example.com");
        }

        @Test
        @DisplayName("UserNotFoundExceptionで404を返す")
        void returnsNotFoundOnUserNotFound() {
            doThrow(UserNotFoundException.builder().message("not found").build())
                .when(cognitoAuthService).forgotPassword(anyString());

            ResponseEntity<?> response = controller.forgotPassword(Map.of("email", "unknown@example.com"));

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NOT_FOUND);
        }
    }

    @Nested
    @DisplayName("confirmForgotPassword")
    class ConfirmForgotPassword {

        @Test
        @DisplayName("パスワードリセット成功で200を返す")
        void returnsOkOnSuccess() {
            ForgotPasswordForm form = new ForgotPasswordForm("test@example.com", "123456", "newPassword123");

            ResponseEntity<?> response = controller.confirmForgotPassword(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(cognitoAuthService).confirmForgotPassword("test@example.com", "123456", "newPassword123");
        }

        @Test
        @DisplayName("CodeMismatchExceptionで400を返す")
        void returnsBadRequestOnCodeMismatch() {
            ForgotPasswordForm form = new ForgotPasswordForm("test@example.com", "wrong", "newPassword123");
            doThrow(CodeMismatchException.builder().message("mismatch").build())
                .when(cognitoAuthService).confirmForgotPassword(anyString(), anyString(), anyString());

            ResponseEntity<?> response = controller.confirmForgotPassword(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }

        @Test
        @DisplayName("ExpiredCodeExceptionで410を返す")
        void returnsGoneOnExpiredCode() {
            ForgotPasswordForm form = new ForgotPasswordForm("test@example.com", "expired", "newPassword123");
            doThrow(ExpiredCodeException.builder().message("expired").build())
                .when(cognitoAuthService).confirmForgotPassword(anyString(), anyString(), anyString());

            ResponseEntity<?> response = controller.confirmForgotPassword(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.GONE);
        }

        @Test
        @DisplayName("InvalidPasswordExceptionで400を返す")
        void returnsBadRequestOnInvalidPassword() {
            ForgotPasswordForm form = new ForgotPasswordForm("test@example.com", "123456", "weak");
            doThrow(InvalidPasswordException.builder().message("invalid").build())
                .when(cognitoAuthService).confirmForgotPassword(anyString(), anyString(), anyString());

            ResponseEntity<?> response = controller.confirmForgotPassword(form);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.BAD_REQUEST);
        }
    }

    @Nested
    @DisplayName("login")
    class Login {

        @Test
        @DisplayName("ログイン成功で200を返す")
        void returnsOkOnSuccess() {
            LoginForm form = new LoginForm("test@example.com", "password123");
            CognitoLoginUseCase.Result result = new CognitoLoginUseCase.Result(
                    "access-token", "refresh-token", "test@example.com", "cognito-user");
            when(cognitoLoginUseCase.execute("test@example.com", "password123")).thenReturn(result);

            ResponseEntity<?> response = controller.login(form, httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(authCookieService).setAuthCookies(httpResponse,
                    "access-token", "refresh-token", "test@example.com", "cognito-user");
        }

        @Test
        @DisplayName("IllegalStateExceptionで401を返す")
        void returnsUnauthorizedOnIllegalState() {
            LoginForm form = new LoginForm("test@example.com", "password123");
            when(cognitoLoginUseCase.execute(anyString(), anyString()))
                    .thenThrow(new IllegalStateException("IDトークンのデコードに失敗"));

            ResponseEntity<?> response = controller.login(form, httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }

        @Test
        @DisplayName("RuntimeExceptionで500を返し内部エラー情報を漏洩しない")
        @SuppressWarnings("unchecked")
        void returnsInternalServerErrorOnRuntimeException() {
            LoginForm form = new LoginForm("test@example.com", "wrong");
            when(cognitoLoginUseCase.execute(anyString(), anyString()))
                    .thenThrow(new RuntimeException("AWS SDK internal timeout error"));

            ResponseEntity<?> response = controller.login(form, httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.INTERNAL_SERVER_ERROR);
            Map<String, String> body = (Map<String, String>) response.getBody();
            assertThat(body.get("error")).doesNotContain("AWS SDK");
        }
    }

    @Nested
    @DisplayName("logout")
    class Logout {

        @Test
        @DisplayName("ログアウト成功で200を返す")
        void returnsOkOnSuccess() {
            Jwt jwt = mockJwt("sub-123");

            ResponseEntity<?> response = controller.logout(jwt, httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(authCookieService).clearRefreshTokenCookie(httpResponse);
        }

        @Test
        @DisplayName("subが空の場合401を返す")
        void returnsUnauthorizedOnEmptySub() {
            Jwt jwt = mockJwt("");

            ResponseEntity<?> response = controller.logout(jwt, httpResponse);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }

    @Nested
    @DisplayName("me")
    class Me {

        @Test
        @DisplayName("認証ユーザー情報を返す")
        void returnsUserInfo() {
            Jwt jwt = mockJwt("sub-123");
            User user = new User();
            user.setId(10);
            when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

            ResponseEntity<?> response = controller.me(jwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        }

        @Test
        @DisplayName("JWTがnullの場合401を返す（/api/auth/**はpermitAllのため）")
        void returnsUnauthorizedOnNullJwt() {
            ResponseEntity<?> response = controller.me(null);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.UNAUTHORIZED);
        }
    }
}
