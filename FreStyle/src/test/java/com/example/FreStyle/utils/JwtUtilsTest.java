package com.example.FreStyle.utils;

import static org.junit.jupiter.api.Assertions.*;

import java.util.Optional;

import org.junit.jupiter.api.Test;

import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.crypto.MACSigner;
import com.nimbusds.jwt.JWTClaimsSet;
import com.nimbusds.jwt.SignedJWT;

class JwtUtilsTest {

    private String createTestJwt(String subject, String issuer) throws Exception {
        // テスト用のHMAC秘密鍵（256ビット以上）
        byte[] secret = new byte[32];
        for (int i = 0; i < secret.length; i++) {
            secret[i] = (byte) i;
        }

        JWTClaimsSet claims = new JWTClaimsSet.Builder()
                .subject(subject)
                .issuer(issuer)
                .build();

        SignedJWT signedJWT = new SignedJWT(
                new JWSHeader(JWSAlgorithm.HS256),
                claims);
        signedJWT.sign(new MACSigner(secret));

        return signedJWT.serialize();
    }

    @Test
    void 正常なJWTをデコードできる() throws Exception {
        String token = createTestJwt("user-123", "https://cognito-idp.ap-northeast-1.amazonaws.com");

        Optional<JWTClaimsSet> result = JwtUtils.decode(token);

        assertTrue(result.isPresent());
        assertEquals("user-123", result.get().getSubject());
        assertEquals("https://cognito-idp.ap-northeast-1.amazonaws.com", result.get().getIssuer());
    }

    @Test
    void 無効なトークンの場合はemptyを返す() {
        Optional<JWTClaimsSet> result = JwtUtils.decode("invalid-token-string");

        assertTrue(result.isEmpty());
    }

    @Test
    void 空文字列の場合はemptyを返す() {
        Optional<JWTClaimsSet> result = JwtUtils.decode("");

        assertTrue(result.isEmpty());
    }

    @Test
    void subjectが設定されたJWTを正しくデコードできる() throws Exception {
        String token = createTestJwt("cognito-sub-abc", null);

        Optional<JWTClaimsSet> result = JwtUtils.decode(token);

        assertTrue(result.isPresent());
        assertEquals("cognito-sub-abc", result.get().getSubject());
    }
}
