package com.example.FreStyle.controller;

import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.AddScenarioBookmarkUseCase;
import com.example.FreStyle.usecase.GetUserBookmarksUseCase;
import com.example.FreStyle.usecase.RemoveScenarioBookmarkUseCase;

import lombok.RequiredArgsConstructor;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/bookmarks")
public class ScenarioBookmarkController {

    private static final Logger logger = LoggerFactory.getLogger(ScenarioBookmarkController.class);

    private final GetUserBookmarksUseCase getUserBookmarksUseCase;
    private final AddScenarioBookmarkUseCase addScenarioBookmarkUseCase;
    private final RemoveScenarioBookmarkUseCase removeScenarioBookmarkUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<Integer>> getBookmarks(@AuthenticationPrincipal Jwt jwt) {
        logger.info("========== GET /api/bookmarks ==========");
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        List<Integer> bookmarkedIds = getUserBookmarksUseCase.execute(user.getId());
        logger.info("ブックマーク一覧取得成功 - 件数: {}", bookmarkedIds.size());
        return ResponseEntity.ok(bookmarkedIds);
    }

    @PostMapping("/{scenarioId}")
    public ResponseEntity<Void> addBookmark(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer scenarioId
    ) {
        logger.info("========== POST /api/bookmarks/{} ==========", scenarioId);
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        addScenarioBookmarkUseCase.execute(user, scenarioId);
        logger.info("ブックマーク追加成功 - scenarioId: {}", scenarioId);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/{scenarioId}")
    public ResponseEntity<Void> removeBookmark(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer scenarioId
    ) {
        logger.info("========== DELETE /api/bookmarks/{} ==========", scenarioId);
        String sub = jwt.getSubject();
        User user = userIdentityService.findUserBySub(sub);
        removeScenarioBookmarkUseCase.execute(user.getId(), scenarioId);
        logger.info("ブックマーク削除成功 - scenarioId: {}", scenarioId);
        return ResponseEntity.ok().build();
    }
}
