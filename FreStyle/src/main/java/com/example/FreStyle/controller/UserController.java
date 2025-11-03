package com.example.FreStyle.controller;

import java.util.Map;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.LoginUserDto;
import com.example.FreStyle.service.UserService;

@RestController
@RequestMapping("/api/user")
public class UserController {
  
  private final UserService userService;
  
  public UserController(UserService userService) {
    this.userService = userService;
  }
  
    @GetMapping("/me")
    public ResponseEntity<?> getUserInfo(@AuthenticationPrincipal Jwt jwt) {
      String sub = jwt.getSubject();
        if (sub == null || sub.isEmpty()) {
          return ResponseEntity.status(HttpStatus.UNAUTHORIZED).body(Map.of("error", "未ログイン"));
        }
        
        LoginUserDto loginUserDto =  userService.findLoginUserByCognitoSub(sub);
        
        return ResponseEntity.ok().body(loginUserDto);
        
    }
}
