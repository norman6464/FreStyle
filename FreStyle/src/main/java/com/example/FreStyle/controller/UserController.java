package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

// React（パブリッククライアント）側でid_token、access_tokenを持っている場合はここで本人か検証をする必要がある

@RestController
@RequestMapping("/api/user")
public class UserController {
  
    @GetMapping("/me")
    public Map<String, Object> getUserInfo(@AuthenticationPrincipal Jwt jwt) {
      return Map.of(
        "sub", jwt.getSubject(),
        "email", jwt.getClaim("email"),
        "name",jwt.getClaim("name")
      ); 
    }
}
