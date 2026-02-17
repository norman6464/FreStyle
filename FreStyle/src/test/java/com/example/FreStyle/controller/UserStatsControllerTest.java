package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.UserStatsDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import com.example.FreStyle.usecase.GetUserStatsUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("UserStatsController テスト")
class UserStatsControllerTest {

    @Mock
    private GetUserStatsUseCase getUserStatsUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private UserService userService;

    @Mock
    private Jwt jwt;

    @InjectMocks
    private UserStatsController userStatsController;

    private User currentUser;

    @BeforeEach
    void setUp() {
        currentUser = new User();
        currentUser.setId(1);
        currentUser.setName("テストユーザー");
    }

    @Test
    @DisplayName("自分の統計を取得できる")
    void getMyStats_returnsStats() {
        when(jwt.getSubject()).thenReturn("sub-123");
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(currentUser);

        UserStatsDto stats = new UserStatsDto(10, 5, 20, 15, 75.5);
        when(getUserStatsUseCase.execute(1)).thenReturn(stats);

        ResponseEntity<UserStatsDto> response = userStatsController.getMyStats(jwt);

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().totalSessions()).isEqualTo(10);
        assertThat(response.getBody().averageScore()).isEqualTo(75.5);
    }

    @Test
    @DisplayName("指定ユーザーの統計を取得できる")
    void getUserStats_returnsStats() {
        User targetUser = new User();
        targetUser.setId(2);
        when(userService.findUserById(2)).thenReturn(targetUser);

        UserStatsDto stats = new UserStatsDto(5, 3, 8, 4, 60.0);
        when(getUserStatsUseCase.execute(2)).thenReturn(stats);

        ResponseEntity<UserStatsDto> response = userStatsController.getUserStats(2);

        assertThat(response.getStatusCode().value()).isEqualTo(200);
        assertThat(response.getBody()).isNotNull();
        assertThat(response.getBody().totalSessions()).isEqualTo(5);
        assertThat(response.getBody().followerCount()).isEqualTo(8);
    }
}
