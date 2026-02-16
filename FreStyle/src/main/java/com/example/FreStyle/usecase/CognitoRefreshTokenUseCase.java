package com.example.FreStyle.usecase;

import com.example.FreStyle.entity.AccessToken;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.CognitoAuthService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;

@Service
@RequiredArgsConstructor
@Slf4j
public class CognitoRefreshTokenUseCase {

    private final AccessTokenService accessTokenService;
    private final CognitoAuthService cognitoAuthService;

    public record Result(String newAccessToken) {}

    public Result execute(String refreshToken, String username) {
        log.info("CognitoRefreshTokenUseCase: トークン更新処理開始");

        AccessToken accessTokenEntity = accessTokenService.findAccessTokenByRefreshToken(refreshToken);
        Map<String, String> tokens = cognitoAuthService.refreshAccessToken(refreshToken, username);

        accessTokenService.updateTokens(accessTokenEntity, tokens.get("accessToken"));

        log.info("CognitoRefreshTokenUseCase: トークン更新処理完了");
        return new Result(tokens.get("accessToken"));
    }
}
