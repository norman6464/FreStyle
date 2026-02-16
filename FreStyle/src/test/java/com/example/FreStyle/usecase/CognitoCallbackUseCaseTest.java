package com.example.FreStyle.usecase;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.MockedStatic;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.core.publisher.Mono;

import java.util.Map;
import java.util.Optional;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class CognitoCallbackUseCaseTest {

    @Mock private WebClient webClient;
    @Mock private UserService userService;
    @Mock private AccessTokenService accessTokenService;

    @InjectMocks
    private CognitoCallbackUseCase cognitoCallbackUseCase;

    @BeforeEach
    void setUp() {
        ReflectionTestUtils.setField(cognitoCallbackUseCase, "clientId", "test-client-id");
        ReflectionTestUtils.setField(cognitoCallbackUseCase, "clientSecret", "test-client-secret");
        ReflectionTestUtils.setField(cognitoCallbackUseCase, "redirectUri", "https://example.com/callback");
        ReflectionTestUtils.setField(cognitoCallbackUseCase, "tokenUri", "https://cognito.example.com/oauth2/token");
    }

    @SuppressWarnings("unchecked")
    private void setupWebClientMock(Mono<?> responseMono) {
        WebClient.RequestBodyUriSpec requestBodyUriSpec = mock(WebClient.RequestBodyUriSpec.class);
        WebClient.RequestBodySpec requestBodySpec = mock(WebClient.RequestBodySpec.class);
        WebClient.RequestHeadersSpec requestHeadersSpec = mock(WebClient.RequestHeadersSpec.class);
        WebClient.ResponseSpec responseSpec = mock(WebClient.ResponseSpec.class);

        when(webClient.post()).thenReturn(requestBodyUriSpec);
        when(requestBodyUriSpec.uri(anyString())).thenReturn(requestBodySpec);
        when(requestBodySpec.header(anyString(), anyString())).thenReturn(requestBodySpec);
        doReturn(requestHeadersSpec).when(requestBodySpec).body(any());
        when(requestHeadersSpec.retrieve()).thenReturn(responseSpec);
        doReturn(responseMono).when(responseSpec).bodyToMono(any(ParameterizedTypeReference.class));
    }

    @Test
    @DisplayName("OIDCコールバック成功時にユーザー登録・トークン保存・Result返却が行われる")
    void returnsResultOnSuccess() throws Exception {
        User user = new User();
        user.setId(1);

        Map<String, Object> tokenResponse = Map.of(
                "id_token", "id-token-value",
                "access_token", "access-token-value",
                "refresh_token", "refresh-token-value");

        setupWebClientMock(Mono.just(tokenResponse));

        JWTClaimsSet claims = new JWTClaimsSet.Builder()
                .subject("sub-123")
                .claim("name", "テストユーザー")
                .claim("email", "test@example.com")
                .claim("cognito:username", "google_12345")
                .claim("identities", "google")
                .build();

        when(userService.registerUserOIDC("テストユーザー", "test@example.com", "google", "sub-123"))
                .thenReturn(user);

        try (MockedStatic<JwtUtils> jwtUtilsMock = mockStatic(JwtUtils.class)) {
            jwtUtilsMock.when(() -> JwtUtils.decode("id-token-value"))
                    .thenReturn(Optional.of(claims));

            CognitoCallbackUseCase.Result result = cognitoCallbackUseCase.execute("auth-code-123");

            assertThat(result.accessToken()).isEqualTo("access-token-value");
            assertThat(result.refreshToken()).isEqualTo("refresh-token-value");
            assertThat(result.email()).isEqualTo("test@example.com");
            assertThat(result.cognitoUsername()).isEqualTo("google_12345");

            verify(accessTokenService).saveTokens(user, "access-token-value", "refresh-token-value");
        }
    }

    @Test
    @DisplayName("トークンレスポンスがnullの場合に例外をスローする")
    void throwsWhenTokenResponseIsNull() {
        setupWebClientMock(Mono.empty());

        assertThatThrownBy(() -> cognitoCallbackUseCase.execute("auth-code-123"))
                .isInstanceOf(IllegalStateException.class)
                .hasMessage("トークン取得に失敗しました。");
    }

    @Test
    @DisplayName("IDトークンのデコード失敗時に例外をスローする")
    void throwsWhenIdTokenDecodeFails() {
        Map<String, Object> tokenResponse = Map.of(
                "id_token", "invalid-token",
                "access_token", "access-token",
                "refresh_token", "refresh-token");

        setupWebClientMock(Mono.just(tokenResponse));

        try (MockedStatic<JwtUtils> jwtUtilsMock = mockStatic(JwtUtils.class)) {
            jwtUtilsMock.when(() -> JwtUtils.decode("invalid-token"))
                    .thenReturn(Optional.empty());

            assertThatThrownBy(() -> cognitoCallbackUseCase.execute("auth-code-123"))
                    .isInstanceOf(IllegalStateException.class)
                    .hasMessage("無効なIDトークンです。");
        }
    }

    @Test
    @DisplayName("cognito:usernameがない場合はsubをcognitoUsernameとして使用する")
    void usesSubAsCognitoUsernameWhenAbsent() throws Exception {
        User user = new User();
        user.setId(1);

        Map<String, Object> tokenResponse = Map.of(
                "id_token", "id-token-value",
                "access_token", "access-token-value",
                "refresh_token", "refresh-token-value");

        setupWebClientMock(Mono.just(tokenResponse));

        JWTClaimsSet claims = new JWTClaimsSet.Builder()
                .subject("sub-456")
                .claim("name", "テストユーザー")
                .claim("email", "test@example.com")
                .build();

        when(userService.registerUserOIDC("テストユーザー", "test@example.com", "cognito", "sub-456"))
                .thenReturn(user);

        try (MockedStatic<JwtUtils> jwtUtilsMock = mockStatic(JwtUtils.class)) {
            jwtUtilsMock.when(() -> JwtUtils.decode("id-token-value"))
                    .thenReturn(Optional.of(claims));

            CognitoCallbackUseCase.Result result = cognitoCallbackUseCase.execute("auth-code-123");

            assertThat(result.cognitoUsername()).isEqualTo("sub-456");
        }
    }
}
