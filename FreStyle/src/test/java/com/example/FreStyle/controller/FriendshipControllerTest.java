package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.usecase.CheckFollowStatusUseCase;
import com.example.FreStyle.usecase.FollowUserUseCase;
import com.example.FreStyle.usecase.GetFollowersUseCase;
import com.example.FreStyle.usecase.GetFollowingUseCase;
import com.example.FreStyle.usecase.UnfollowUserUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("FriendshipController テスト")
class FriendshipControllerTest {

    @Mock
    private FollowUserUseCase followUserUseCase;

    @Mock
    private UnfollowUserUseCase unfollowUserUseCase;

    @Mock
    private GetFollowersUseCase getFollowersUseCase;

    @Mock
    private GetFollowingUseCase getFollowingUseCase;

    @Mock
    private CheckFollowStatusUseCase checkFollowStatusUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private UserService userService;

    @InjectMocks
    private FriendshipController friendshipController;

    private Jwt mockJwt;
    private User testUser;
    private User targetUser;

    @BeforeEach
    void setUp() {
        mockJwt = mock(Jwt.class);
        when(mockJwt.getSubject()).thenReturn("sub-123");

        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");

        targetUser = new User();
        targetUser.setId(2);
        targetUser.setName("ターゲット");

        when(userIdentityService.findUserBySub("sub-123")).thenReturn(testUser);
    }

    @Nested
    @DisplayName("followUser")
    class FollowUser {

        @Test
        @DisplayName("ユーザーをフォローできる")
        void followsUser() {
            when(userService.findUserById(2)).thenReturn(targetUser);
            FriendshipDto dto = new FriendshipDto(1, 2, "ターゲット", null, null, false, "2026-02-17T00:00:00");
            when(followUserUseCase.execute(testUser, targetUser)).thenReturn(dto);

            ResponseEntity<FriendshipDto> response = friendshipController.followUser(mockJwt, 2);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().getUserId()).isEqualTo(2);
        }
    }

    @Nested
    @DisplayName("unfollowUser")
    class UnfollowUser {

        @Test
        @DisplayName("フォローを解除できる")
        void unfollowsUser() {
            ResponseEntity<Void> response = friendshipController.unfollowUser(mockJwt, 2);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.NO_CONTENT);
            verify(unfollowUserUseCase).execute(testUser, 2);
        }
    }

    @Nested
    @DisplayName("getFollowing")
    class GetFollowing {

        @Test
        @DisplayName("フォロー中一覧を取得できる")
        void returnsFollowingList() {
            FriendshipDto dto = new FriendshipDto(1, 2, "ターゲット", null, null, false, "2026-02-17T00:00:00");
            when(getFollowingUseCase.execute(1)).thenReturn(List.of(dto));

            ResponseEntity<List<FriendshipDto>> response = friendshipController.getFollowing(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).hasSize(1);
        }
    }

    @Nested
    @DisplayName("getFollowers")
    class GetFollowers {

        @Test
        @DisplayName("フォロワー一覧を取得できる")
        void returnsFollowerList() {
            FriendshipDto dto = new FriendshipDto(1, 2, "フォロワー", null, null, true, "2026-02-17T00:00:00");
            when(getFollowersUseCase.execute(1)).thenReturn(List.of(dto));

            ResponseEntity<List<FriendshipDto>> response = friendshipController.getFollowers(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).hasSize(1);
        }
    }

    @Nested
    @DisplayName("checkFollowStatus")
    class CheckFollowStatus {

        @Test
        @DisplayName("フォロー状態を確認できる")
        void returnsFollowStatus() {
            Map<String, Object> status = Map.of("isFollowing", true, "isFollowedBy", false, "isMutual", false);
            when(checkFollowStatusUseCase.execute(1, 2)).thenReturn(status);

            ResponseEntity<Map<String, Object>> response = friendshipController.checkFollowStatus(mockJwt, 2);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody().get("isFollowing")).isEqualTo(true);
        }
    }
}
