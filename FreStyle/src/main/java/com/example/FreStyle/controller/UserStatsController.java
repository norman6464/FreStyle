package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.UserStatsDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.usecase.GetUserStatsUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/users")
@Slf4j
public class UserStatsController {

    private final GetUserStatsUseCase getUserStatsUseCase;
    private final UserIdentityService userIdentityService;
    private final UserService userService;

    @GetMapping("/me/stats")
    public ResponseEntity<UserStatsDto> getMyStats(@AuthenticationPrincipal Jwt jwt) {
        User user = userIdentityService.findUserBySub(jwt.getSubject());
        log.info("ユーザー統計取得: userId={}", user.getId());
        UserStatsDto stats = getUserStatsUseCase.execute(user.getId());
        return ResponseEntity.ok(stats);
    }

    @GetMapping("/{userId}/stats")
    public ResponseEntity<UserStatsDto> getUserStats(@PathVariable Integer userId) {
        userService.findUserById(userId);
        log.info("ユーザー統計取得: userId={}", userId);
        UserStatsDto stats = getUserStatsUseCase.execute(userId);
        return ResponseEntity.ok(stats);
    }
}
