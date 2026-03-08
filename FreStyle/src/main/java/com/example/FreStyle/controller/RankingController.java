package com.example.FreStyle.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.RankingDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetRankingUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/ranking")
@Slf4j
public class RankingController {

    private final UserIdentityService userIdentityService;
    private final GetRankingUseCase getRankingUseCase;

    @GetMapping
    public ResponseEntity<RankingDto> getRanking(
            @AuthenticationPrincipal Jwt jwt,
            @RequestParam(defaultValue = "weekly") String period
    ) {
        log.info("========== GET /api/ranking?period={} ==========", period);

        User user = userIdentityService.findUserBySub(jwt.getSubject());
        RankingDto ranking = getRankingUseCase.execute(period, user.getId());

        log.info("ランキング取得成功 - userId: {}, 件数: {}", user.getId(), ranking.entries().size());

        return ResponseEntity.ok(ranking);
    }
}
