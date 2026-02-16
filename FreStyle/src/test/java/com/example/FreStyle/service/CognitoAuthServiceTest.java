package com.example.FreStyle.service;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.springframework.test.util.ReflectionTestUtils;

import software.amazon.awssdk.services.cognitoidentityprovider.CognitoIdentityProviderClient;
import software.amazon.awssdk.services.cognitoidentityprovider.model.*;

class CognitoAuthServiceTest {

    private CognitoIdentityProviderClient cognitoClient;
    private CognitoAuthService service;

    @BeforeEach
    void setUp() {
        cognitoClient = mock(CognitoIdentityProviderClient.class);
        service = new CognitoAuthService("dummyAccessKey", "dummySecretKey", "ap-northeast-1");
        ReflectionTestUtils.setField(service, "cognitoClient", cognitoClient);
        ReflectionTestUtils.setField(service, "clientId", "testClientId");
        ReflectionTestUtils.setField(service, "clientSecret", "testClientSecret");
    }

    // ============================
    // signUpUser
    // ============================
    @Nested
    class SignUpUserTest {

        @Test
        void 正常にサインアップできる() {
            SignUpResponse response = SignUpResponse.builder()
                    .userConfirmed(false)
                    .build();
            when(cognitoClient.signUp(any(SignUpRequest.class))).thenReturn(response);

            assertDoesNotThrow(() -> service.signUpUser("test@example.com", "Password123!", "テストユーザー"));
            verify(cognitoClient).signUp(any(SignUpRequest.class));
        }

        @Test
        void 既存メールアドレスでUsernameExistsExceptionがそのまま伝搬する() {
            when(cognitoClient.signUp(any(SignUpRequest.class)))
                    .thenThrow(UsernameExistsException.builder().message("exists").build());

            assertThrows(UsernameExistsException.class,
                    () -> service.signUpUser("existing@example.com", "Password123!", "テスト"));
        }

        @Test
        void 不正なパスワードでInvalidPasswordExceptionがそのまま伝搬する() {
            when(cognitoClient.signUp(any(SignUpRequest.class)))
                    .thenThrow(InvalidPasswordException.builder().message("weak").build());

            assertThrows(InvalidPasswordException.class,
                    () -> service.signUpUser("test@example.com", "weak", "テスト"));
        }

        @Test
        void 確認済みユーザーの場合も正常終了する() {
            SignUpResponse response = SignUpResponse.builder()
                    .userConfirmed(true)
                    .build();
            when(cognitoClient.signUp(any(SignUpRequest.class))).thenReturn(response);

            assertDoesNotThrow(() -> service.signUpUser("confirmed@example.com", "Password123!", "確認済み"));
        }
    }

    // ============================
    // confirmUserSignup
    // ============================
    @Nested
    class ConfirmUserSignupTest {

        @Test
        void 正常に確認できる() {
            when(cognitoClient.confirmSignUp(any(ConfirmSignUpRequest.class)))
                    .thenReturn(ConfirmSignUpResponse.builder().build());

            assertDoesNotThrow(() -> service.confirmUserSignup("test@example.com", "123456"));
            verify(cognitoClient).confirmSignUp(any(ConfirmSignUpRequest.class));
        }

        @Test
        void ユーザー未存在でUserNotFoundExceptionがそのまま伝搬する() {
            when(cognitoClient.confirmSignUp(any(ConfirmSignUpRequest.class)))
                    .thenThrow(UserNotFoundException.builder().message("not found").build());

            assertThrows(UserNotFoundException.class,
                    () -> service.confirmUserSignup("unknown@example.com", "123456"));
        }

        @Test
        void 確認コード不一致でCodeMismatchExceptionがそのまま伝搬する() {
            when(cognitoClient.confirmSignUp(any(ConfirmSignUpRequest.class)))
                    .thenThrow(CodeMismatchException.builder().message("mismatch").build());

            assertThrows(CodeMismatchException.class,
                    () -> service.confirmUserSignup("test@example.com", "000000"));
        }

        @Test
        void 確認コード期限切れでExpiredCodeExceptionがそのまま伝搬する() {
            when(cognitoClient.confirmSignUp(any(ConfirmSignUpRequest.class)))
                    .thenThrow(ExpiredCodeException.builder().message("expired").build());

            assertThrows(ExpiredCodeException.class,
                    () -> service.confirmUserSignup("test@example.com", "123456"));
        }
    }

    // ============================
    // login
    // ============================
    @Nested
    class LoginTest {

        @Test
        void 正常にログインできる() {
            AuthenticationResultType authResult = AuthenticationResultType.builder()
                    .accessToken("access-token")
                    .idToken("id-token")
                    .refreshToken("refresh-token")
                    .build();
            InitiateAuthResponse response = InitiateAuthResponse.builder()
                    .authenticationResult(authResult)
                    .build();
            when(cognitoClient.initiateAuth(any(InitiateAuthRequest.class))).thenReturn(response);

            Map<String, String> tokens = service.login("test@example.com", "Password123!");

            assertEquals("access-token", tokens.get("accessToken"));
            assertEquals("id-token", tokens.get("idToken"));
            assertEquals("refresh-token", tokens.get("refreshToken"));
        }

        @Test
        void 認証失敗でNotAuthorizedExceptionがそのまま伝搬する() {
            when(cognitoClient.initiateAuth(any(InitiateAuthRequest.class)))
                    .thenThrow(NotAuthorizedException.builder().message("bad creds").build());

            assertThrows(NotAuthorizedException.class,
                    () -> service.login("test@example.com", "WrongPassword"));
        }

        @Test
        void メール未確認でUserNotConfirmedExceptionがそのまま伝搬する() {
            when(cognitoClient.initiateAuth(any(InitiateAuthRequest.class)))
                    .thenThrow(UserNotConfirmedException.builder().message("not confirmed").build());

            assertThrows(UserNotConfirmedException.class,
                    () -> service.login("unconfirmed@example.com", "Password123!"));
        }
    }

    // ============================
    // forgotPassword
    // ============================
    @Nested
    class ForgotPasswordTest {

        @Test
        void 正常にパスワードリセットリクエストできる() {
            when(cognitoClient.forgotPassword(any(ForgotPasswordRequest.class)))
                    .thenReturn(ForgotPasswordResponse.builder().build());

            assertDoesNotThrow(() -> service.forgotPassword("test@example.com"));
            verify(cognitoClient).forgotPassword(any(ForgotPasswordRequest.class));
        }

        @Test
        void ユーザー未存在でUserNotFoundExceptionがそのまま伝搬する() {
            when(cognitoClient.forgotPassword(any(ForgotPasswordRequest.class)))
                    .thenThrow(UserNotFoundException.builder().message("not found").build());

            assertThrows(UserNotFoundException.class,
                    () -> service.forgotPassword("unknown@example.com"));
        }
    }

    // ============================
    // confirmForgotPassword
    // ============================
    @Nested
    class ConfirmForgotPasswordTest {

        @Test
        void 正常にパスワードリセットできる() {
            when(cognitoClient.confirmForgotPassword(any(ConfirmForgotPasswordRequest.class)))
                    .thenReturn(ConfirmForgotPasswordResponse.builder().build());

            assertDoesNotThrow(() -> service.confirmForgotPassword("test@example.com", "123456", "NewPassword123!"));
            verify(cognitoClient).confirmForgotPassword(any(ConfirmForgotPasswordRequest.class));
        }

        @Test
        void 確認コード不一致でCodeMismatchExceptionがそのまま伝搬する() {
            when(cognitoClient.confirmForgotPassword(any(ConfirmForgotPasswordRequest.class)))
                    .thenThrow(CodeMismatchException.builder().message("mismatch").build());

            assertThrows(CodeMismatchException.class,
                    () -> service.confirmForgotPassword("test@example.com", "000000", "NewPassword123!"));
        }

        @Test
        void 確認コード期限切れでExpiredCodeExceptionがそのまま伝搬する() {
            when(cognitoClient.confirmForgotPassword(any(ConfirmForgotPasswordRequest.class)))
                    .thenThrow(ExpiredCodeException.builder().message("expired").build());

            assertThrows(ExpiredCodeException.class,
                    () -> service.confirmForgotPassword("test@example.com", "123456", "NewPassword123!"));
        }
    }

    // ============================
    // updateUserProfile
    // ============================
    @Nested
    class UpdateUserProfileTest {

        @Test
        void 正常にプロフィール更新できる() {
            when(cognitoClient.updateUserAttributes(any(UpdateUserAttributesRequest.class)))
                    .thenReturn(UpdateUserAttributesResponse.builder().build());

            assertDoesNotThrow(() -> service.updateUserProfile("access-token", "新しい名前"));
            verify(cognitoClient).updateUserAttributes(any(UpdateUserAttributesRequest.class));
        }

        @Test
        void 空の名前でRuntimeException() {
            RuntimeException ex = assertThrows(RuntimeException.class,
                    () -> service.updateUserProfile("access-token", ""));
            assertEquals("proceed update profile error", ex.getMessage());
        }

        @Test
        void nullの名前でRuntimeException() {
            RuntimeException ex = assertThrows(RuntimeException.class,
                    () -> service.updateUserProfile("access-token", null));
            assertEquals("proceed update profile error", ex.getMessage());
        }

        @Test
        void 認証失敗でNotAuthorizedException() {
            when(cognitoClient.updateUserAttributes(any(UpdateUserAttributesRequest.class)))
                    .thenThrow(NotAuthorizedException.builder().message("not authorized").build());

            RuntimeException ex = assertThrows(RuntimeException.class,
                    () -> service.updateUserProfile("invalid-token", "名前"));
            assertEquals("invalid parameter retry login", ex.getMessage());
        }

        @Test
        void InvalidUserPoolConfigurationExceptionでRuntimeException() {
            when(cognitoClient.updateUserAttributes(any(UpdateUserAttributesRequest.class)))
                    .thenThrow(InvalidUserPoolConfigurationException.builder().message("bad config").build());

            RuntimeException ex = assertThrows(RuntimeException.class,
                    () -> service.updateUserProfile("access-token", "名前"));
            assertEquals("user configuration setting", ex.getMessage());
        }

        @Test
        void InvalidParameterExceptionでRuntimeException() {
            when(cognitoClient.updateUserAttributes(any(UpdateUserAttributesRequest.class)))
                    .thenThrow(new java.security.InvalidParameterException("invalid param"));

            RuntimeException ex = assertThrows(RuntimeException.class,
                    () -> service.updateUserProfile("access-token", "名前"));
            assertEquals("invalid parameter", ex.getMessage());
        }
    }

    // ============================
    // refreshAccessToken
    // ============================
    @Nested
    class RefreshAccessTokenTest {

        @Test
        void 正常にトークンを再発行できる() {
            AuthenticationResultType authResult = AuthenticationResultType.builder()
                    .accessToken("new-access-token")
                    .idToken("new-id-token")
                    .build();
            InitiateAuthResponse response = InitiateAuthResponse.builder()
                    .authenticationResult(authResult)
                    .build();
            when(cognitoClient.initiateAuth(any(InitiateAuthRequest.class))).thenReturn(response);

            Map<String, String> tokens = service.refreshAccessToken("valid-refresh-token", "test@example.com");

            assertEquals("new-access-token", tokens.get("accessToken"));
            assertEquals("new-id-token", tokens.get("idToken"));
        }

        @Test
        void 無効なリフレッシュトークンでNotAuthorizedException() {
            when(cognitoClient.initiateAuth(any(InitiateAuthRequest.class)))
                    .thenThrow(NotAuthorizedException.builder().message("invalid token").build());

            RuntimeException ex = assertThrows(RuntimeException.class,
                    () -> service.refreshAccessToken("invalid-token", "test@example.com"));
            assertEquals("リフレッシュトークンが無効です。再ログインしてください。", ex.getMessage());
        }
    }
}
