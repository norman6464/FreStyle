package com.example.FreStyle.service;


import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.AccessToken;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AccessTokenRepository;

@Service
public class AccessTokenService {

    private final AccessTokenRepository accessTokenRepository;

    public AccessTokenService(AccessTokenRepository accessTokenRepository) {
        this.accessTokenRepository = accessTokenRepository;
    }

    @Transactional
    public void saveTokens(
            User user,
            String accessToken,
            String refreshToken
    ) {
        AccessToken token = new AccessToken();
        token.setAccessToken(accessToken);
        token.setUser(user);
        token.setRefreshToken(refreshToken);
        token.setRevoked(false);

        accessTokenRepository.save(token);
    }
}
