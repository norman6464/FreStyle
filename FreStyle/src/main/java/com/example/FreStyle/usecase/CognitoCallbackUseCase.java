package com.example.FreStyle.usecase;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.core.ParameterizedTypeReference;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Service;
import org.springframework.util.LinkedMultiValueMap;
import org.springframework.util.MultiValueMap;
import org.springframework.web.reactive.function.BodyInserters;
import org.springframework.web.reactive.function.client.WebClient;

import java.nio.charset.StandardCharsets;
import java.util.Base64;
import java.util.Map;
import java.util.Optional;

@Service
@RequiredArgsConstructor
@Slf4j
public class CognitoCallbackUseCase {

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
    private final AccessTokenService accessTokenService;

    public record Result(String accessToken, String refreshToken, String email, String cognitoUsername) {}

    public Result execute(String code) {
        log.info("CognitoCallbackUseCase: OIDCコールバック処理開始");

        Map<String, Object> tokenResponse = exchangeCodeForTokens(code);

        if (tokenResponse == null) {
            throw new IllegalStateException("トークン取得に失敗しました。");
        }

        String idToken = (String) tokenResponse.get("id_token");
        String accessToken = (String) tokenResponse.get("access_token");
        String refreshToken = (String) tokenResponse.get("refresh_token");

        Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);
        if (claimsOpt.isEmpty()) {
            throw new IllegalStateException("無効なIDトークンです。");
        }

        try {
            JWTClaimsSet claims = claimsOpt.get();
            String name = claims.getStringClaim("name");
            String email = claims.getStringClaim("email");
            String sub = claims.getSubject();

            boolean isGoogle = claims.getClaim("identities") != null;
            String provider = isGoogle ? "google" : "cognito";

            User user = userService.registerUserOIDC(name, email, provider, sub);

            Object cognitoUsernameObj = claims.getClaim("cognito:username");
            String cognitoUsername = cognitoUsernameObj != null ? cognitoUsernameObj.toString() : sub;

            accessTokenService.saveTokens(user, accessToken, refreshToken);

            log.info("CognitoCallbackUseCase: OIDCコールバック処理完了 - userId: {}", user.getId());
            return new Result(accessToken, refreshToken, email, cognitoUsername);

        } catch (Exception e) {
            log.error("CognitoCallbackUseCase: OIDCコールバック処理中にエラー: {}", e.getMessage(), e);
            throw new RuntimeException(e.getMessage(), e);
        }
    }

    private Map<String, Object> exchangeCodeForTokens(String code) {
        String basicAuthValue = Base64.getEncoder()
                .encodeToString((clientId + ":" + clientSecret).getBytes(StandardCharsets.UTF_8));

        MultiValueMap<String, String> formData = new LinkedMultiValueMap<>();
        formData.add("grant_type", "authorization_code");
        formData.add("code", code);
        formData.add("redirect_uri", redirectUri);
        formData.add("client_id", clientId);

        return webClient.post()
                .uri(tokenUri)
                .header(HttpHeaders.CONTENT_TYPE, MediaType.APPLICATION_FORM_URLENCODED_VALUE)
                .header(HttpHeaders.AUTHORIZATION, "Basic " + basicAuthValue)
                .body(BodyInserters.fromFormData(formData))
                .retrieve()
                .bodyToMono(new ParameterizedTypeReference<Map<String, Object>>() {})
                .block();
    }
}
