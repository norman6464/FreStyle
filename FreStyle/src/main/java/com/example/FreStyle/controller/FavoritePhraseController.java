package com.example.FreStyle.controller;

import java.util.List;

import org.springframework.http.ResponseEntity;
import jakarta.validation.Valid;
import jakarta.validation.constraints.NotBlank;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.FavoritePhraseDto;
import com.example.FreStyle.dto.FavoritePhraseSummaryDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.AddFavoritePhraseUseCase;
import com.example.FreStyle.usecase.GetFavoritePhraseSummaryUseCase;
import com.example.FreStyle.usecase.GetUserFavoritePhrasesUseCase;
import com.example.FreStyle.usecase.RemoveFavoritePhraseUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/favorite-phrases")
@Slf4j
public class FavoritePhraseController {

    private final GetUserFavoritePhrasesUseCase getUserFavoritePhrasesUseCase;
    private final AddFavoritePhraseUseCase addFavoritePhraseUseCase;
    private final RemoveFavoritePhraseUseCase removeFavoritePhraseUseCase;
    private final GetFavoritePhraseSummaryUseCase getFavoritePhraseSummaryUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<FavoritePhraseDto>> getFavoritePhrases(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        List<FavoritePhraseDto> phrases = getUserFavoritePhrasesUseCase.execute(user.getId());
        return ResponseEntity.ok(phrases);
    }

    @PostMapping
    public ResponseEntity<Void> addFavoritePhrase(
            @AuthenticationPrincipal Jwt jwt,
            @Valid @RequestBody AddFavoritePhraseRequest request) {
        User user = resolveUser(jwt);
        addFavoritePhraseUseCase.execute(
                user,
                request.originalText(),
                request.rephrasedText(),
                request.pattern());
        return ResponseEntity.ok().build();
    }

    @GetMapping("/summary")
    public ResponseEntity<FavoritePhraseSummaryDto> getSummary(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("お気に入りフレーズサマリー取得: userId={}", user.getId());
        FavoritePhraseSummaryDto summary = getFavoritePhraseSummaryUseCase.execute(user.getId());
        return ResponseEntity.ok(summary);
    }

    @DeleteMapping("/{phraseId}")
    public ResponseEntity<Void> removeFavoritePhrase(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer phraseId) {
        User user = resolveUser(jwt);
        removeFavoritePhraseUseCase.execute(user.getId(), phraseId);
        return ResponseEntity.ok().build();
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }

    record AddFavoritePhraseRequest(@NotBlank String originalText, @NotBlank String rephrasedText, @NotBlank String pattern) {}
}
