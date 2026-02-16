package com.example.FreStyle.usecase;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.MockedStatic;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Map;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class CognitoLoginUseCaseTest {

    @Mock private UserService userService;
    @Mock private CognitoAuthService cognitoAuthService;
    @Mock private UserIdentityService userIdentityService;
    @Mock private AccessTokenService accessTokenService;

    @InjectMocks
    private CognitoLoginUseCase cognitoLoginUseCase;

    @Test
    @DisplayName("ログイン成功時にResult（トークン情報）を返す")
    void returnsResultOnSuccess() throws Exception {
        User user = new User();
        user.setId(1);

        when(cognitoAuthService.login("test@example.com", "password123"))
                .thenReturn(Map.of(
                        "idToken", "id-token-value",
                        "accessToken", "access-token-value",
                        "refreshToken", "refresh-token-value"));
        when(userService.findUserByEmail("test@example.com")).thenReturn(user);

        JWTClaimsSet claims = new JWTClaimsSet.Builder()
                .subject("sub-123")
                .issuer("https://cognito-idp.example.com")
                .claim("cognito:username", "testuser")
                .build();

        try (MockedStatic<JwtUtils> jwtUtilsMock = mockStatic(JwtUtils.class)) {
            jwtUtilsMock.when(() -> JwtUtils.decode("id-token-value"))
                    .thenReturn(Optional.of(claims));

            CognitoLoginUseCase.Result result = cognitoLoginUseCase.execute("test@example.com", "password123");

            assertThat(result.accessToken()).isEqualTo("access-token-value");
            assertThat(result.refreshToken()).isEqualTo("refresh-token-value");
            assertThat(result.email()).isEqualTo("test@example.com");
            assertThat(result.cognitoUsername()).isEqualTo("testuser");

            verify(userService).checkUserIsActive("test@example.com");
            verify(userIdentityService).registerUserIdentity(user, "https://cognito-idp.example.com", "sub-123");
            verify(accessTokenService).saveTokens(user, "access-token-value", "refresh-token-value");
        }
    }

    @Test
    @DisplayName("IDトークンのデコード失敗で例外をスローする")
    void throwsOnInvalidIdToken() {
        when(cognitoAuthService.login("test@example.com", "password123"))
                .thenReturn(Map.of(
                        "idToken", "invalid-token",
                        "accessToken", "access-token",
                        "refreshToken", "refresh-token"));

        try (MockedStatic<JwtUtils> jwtUtilsMock = mockStatic(JwtUtils.class)) {
            jwtUtilsMock.when(() -> JwtUtils.decode("invalid-token"))
                    .thenReturn(Optional.empty());

            assertThatThrownBy(() -> cognitoLoginUseCase.execute("test@example.com", "password123"))
                    .isInstanceOf(IllegalStateException.class)
                    .hasMessage("IDトークンのデコードに失敗しました。");
        }
    }
}
