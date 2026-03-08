package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.WeeklyChallengeDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetCurrentChallengeUseCase;
import com.example.FreStyle.usecase.IncrementChallengeProgressUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/weekly-challenge")
@Slf4j
public class WeeklyChallengeController {

    private final GetCurrentChallengeUseCase getCurrentChallengeUseCase;
    private final IncrementChallengeProgressUseCase incrementChallengeProgressUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<WeeklyChallengeDto> getCurrentChallenge(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("ウィークリーチャレンジ取得: userId={}", user.getId());
        WeeklyChallengeDto dto = getCurrentChallengeUseCase.execute(user.getId());
        return ResponseEntity.ok(dto);
    }

    @PostMapping("/progress")
    public ResponseEntity<WeeklyChallengeDto> incrementProgress(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("ウィークリーチャレンジ進捗インクリメント: userId={}", user.getId());
        WeeklyChallengeDto dto = incrementChallengeProgressUseCase.execute(user.getId());
        return ResponseEntity.ok(dto);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}
