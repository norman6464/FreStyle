package com.example.FreStyle.controller;

import java.util.List;
import java.util.Map;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.usecase.CheckFollowStatusUseCase;
import com.example.FreStyle.usecase.FollowUserUseCase;
import com.example.FreStyle.usecase.GetFollowersUseCase;
import com.example.FreStyle.usecase.GetFollowingUseCase;
import com.example.FreStyle.usecase.UnfollowUserUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/friendships")
@Slf4j
public class FriendshipController {

    private final FollowUserUseCase followUserUseCase;
    private final UnfollowUserUseCase unfollowUserUseCase;
    private final GetFollowersUseCase getFollowersUseCase;
    private final GetFollowingUseCase getFollowingUseCase;
    private final CheckFollowStatusUseCase checkFollowStatusUseCase;
    private final UserIdentityService userIdentityService;
    private final UserService userService;

    @PostMapping("/{userId}/follow")
    public ResponseEntity<FriendshipDto> followUser(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer userId) {
        User currentUser = resolveUser(jwt);
        User targetUser = userService.findUserById(userId);
        log.info("フォロー: userId={} -> targetId={}", currentUser.getId(), userId);
        FriendshipDto result = followUserUseCase.execute(currentUser, targetUser);
        return ResponseEntity.ok(result);
    }

    @DeleteMapping("/{userId}/follow")
    public ResponseEntity<Void> unfollowUser(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer userId) {
        User currentUser = resolveUser(jwt);
        log.info("フォロー解除: userId={} -> targetId={}", currentUser.getId(), userId);
        unfollowUserUseCase.execute(currentUser, userId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping("/following")
    public ResponseEntity<List<FriendshipDto>> getFollowing(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("フォロー中一覧取得: userId={}", user.getId());
        List<FriendshipDto> following = getFollowingUseCase.execute(user.getId());
        return ResponseEntity.ok(following);
    }

    @GetMapping("/followers")
    public ResponseEntity<List<FriendshipDto>> getFollowers(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        log.info("フォロワー一覧取得: userId={}", user.getId());
        List<FriendshipDto> followers = getFollowersUseCase.execute(user.getId());
        return ResponseEntity.ok(followers);
    }

    @GetMapping("/{userId}/status")
    public ResponseEntity<Map<String, Object>> checkFollowStatus(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer userId) {
        User user = resolveUser(jwt);
        Map<String, Object> status = checkFollowStatusUseCase.execute(user.getId(), userId);
        return ResponseEntity.ok(status);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}
