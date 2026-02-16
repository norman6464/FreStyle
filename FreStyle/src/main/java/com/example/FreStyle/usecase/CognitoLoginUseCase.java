package com.example.FreStyle.usecase;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.AccessTokenService;
import com.example.FreStyle.service.CognitoAuthService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.utils.JwtUtils;
import com.nimbusds.jwt.JWTClaimsSet;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.Optional;

@Service
@RequiredArgsConstructor
@Slf4j
public class CognitoLoginUseCase {

    private final UserService userService;
    private final CognitoAuthService cognitoAuthService;
    private final UserIdentityService userIdentityService;
    private final AccessTokenService accessTokenService;

    public record Result(String accessToken, String refreshToken, String email, String cognitoUsername) {}

    public Result execute(String email, String password) {
        log.info("CognitoLoginUseCase: ログイン処理開始 - email: {}", email);

        userService.checkUserIsActive(email);

        Map<String, String> tokens = cognitoAuthService.login(email, password);
        String idToken = tokens.get("idToken");
        String accessToken = tokens.get("accessToken");
        String refreshToken = tokens.get("refreshToken");

        Optional<JWTClaimsSet> claimsOpt = JwtUtils.decode(idToken);
        if (claimsOpt.isEmpty()) {
            throw new IllegalStateException("IDトークンのデコードに失敗しました。");
        }

        JWTClaimsSet claims = claimsOpt.get();
        User user = userService.findUserByEmail(email);
        userIdentityService.registerUserIdentity(user, claims.getIssuer(), claims.getSubject());

        Object cognitoUsernameObj = claims.getClaim("cognito:username");
        String cognitoUsername = cognitoUsernameObj != null ? cognitoUsernameObj.toString() : email;

        accessTokenService.saveTokens(user, accessToken, refreshToken);

        log.info("CognitoLoginUseCase: ログイン処理完了 - userId: {}", user.getId());
        return new Result(accessToken, refreshToken, email, cognitoUsername);
    }
}
