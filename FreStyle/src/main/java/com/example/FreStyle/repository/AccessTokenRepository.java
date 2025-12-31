package com.example.FreStyle.repository;

import org.springframework.data.jpa.repository.JpaRepository;

import com.example.FreStyle.entity.AccessToken;

public interface AccessTokenRepository extends JpaRepository <AccessToken, String> {
     
}
