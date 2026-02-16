package com.example.FreStyle.service;

import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseCookie;
import org.springframework.stereotype.Service;

import jakarta.servlet.http.HttpServletResponse;

@Service
public class AuthCookieService {

    private static final long ACCESS_TOKEN_MAX_AGE = 60L * 60 * 2;
    private static final long PERSISTENT_COOKIE_MAX_AGE = 60L * 60 * 24 * 7;

    public void setAuthCookies(
            HttpServletResponse response,
            String accessToken,
            String refreshToken,
            String email,
            String cognitoUsername) {

        addCookie(response, "ACCESS_TOKEN", accessToken, ACCESS_TOKEN_MAX_AGE, true);
        addCookie(response, "REFRESH_TOKEN", refreshToken, PERSISTENT_COOKIE_MAX_AGE, true);
        addCookie(response, "EMAIL", email, PERSISTENT_COOKIE_MAX_AGE, true);
        addCookie(response, "COGNITO_USERNAME", cognitoUsername, PERSISTENT_COOKIE_MAX_AGE, true);
    }

    public void clearRefreshTokenCookie(HttpServletResponse response) {
        addCookie(response, "REFRESH_TOKEN", null, 0, false);
    }

    private void addCookie(HttpServletResponse response, String name, String value, long maxAge, boolean secure) {
        ResponseCookie cookie = ResponseCookie.from(name, value)
                .httpOnly(true)
                .secure(secure)
                .path("/")
                .maxAge(maxAge)
                .sameSite("None")
                .build();
        response.addHeader(HttpHeaders.SET_COOKIE, cookie.toString());
    }
}
