package com.example.FreStyle.repository;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.AccessToken;
import java.util.Optional;


public interface AccessTokenRepository extends JpaRepository <AccessToken, Integer> {
    AccessToken findByAccessToken(String accessToken);
    Optional<AccessToken> findByRefreshToken(String refreshToken);
}
