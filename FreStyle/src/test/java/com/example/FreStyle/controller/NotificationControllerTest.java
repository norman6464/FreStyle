package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

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

import com.example.FreStyle.dto.NotificationDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetUnreadNotificationCountUseCase;
import com.example.FreStyle.usecase.GetUserNotificationsUseCase;
import com.example.FreStyle.usecase.MarkAllNotificationsAsReadUseCase;
import com.example.FreStyle.usecase.MarkNotificationAsReadUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("NotificationController テスト")
class NotificationControllerTest {

    @Mock
    private GetUserNotificationsUseCase getUserNotificationsUseCase;

    @Mock
    private MarkNotificationAsReadUseCase markNotificationAsReadUseCase;

    @Mock
    private MarkAllNotificationsAsReadUseCase markAllNotificationsAsReadUseCase;

    @Mock
    private GetUnreadNotificationCountUseCase getUnreadNotificationCountUseCase;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private NotificationController notificationController;

    private Jwt mockJwt;
    private User testUser;

    @BeforeEach
    void setUp() {
        mockJwt = mock(Jwt.class);
        when(mockJwt.getSubject()).thenReturn("sub-123");

        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");

        when(userIdentityService.findUserBySub("sub-123")).thenReturn(testUser);
    }

    @Nested
    @DisplayName("getNotifications")
    class GetNotifications {

        @Test
        @DisplayName("通知一覧を取得できる")
        void returnsNotifications() {
            NotificationDto dto = new NotificationDto(1, "NEW_MESSAGE", "新しいメッセージ", "田中さんから", false, 5, "2026-01-15T10:30:00");
            when(getUserNotificationsUseCase.execute(1)).thenReturn(List.of(dto));

            ResponseEntity<List<NotificationDto>> response = notificationController.getNotifications(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).hasSize(1);
            assertThat(response.getBody().get(0).getType()).isEqualTo("NEW_MESSAGE");
        }

        @Test
        @DisplayName("通知がない場合は空リストを返す")
        void returnsEmptyList() {
            when(getUserNotificationsUseCase.execute(1)).thenReturn(List.of());

            ResponseEntity<List<NotificationDto>> response = notificationController.getNotifications(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEmpty();
        }
    }

    @Nested
    @DisplayName("markAsRead")
    class MarkAsRead {

        @Test
        @DisplayName("通知を既読にできる")
        void marksAsRead() {
            ResponseEntity<Void> response = notificationController.markAsRead(mockJwt, 10);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(markNotificationAsReadUseCase).execute(1, 10);
        }
    }

    @Nested
    @DisplayName("markAllAsRead")
    class MarkAllAsRead {

        @Test
        @DisplayName("全通知を既読にできる")
        void marksAllAsRead() {
            ResponseEntity<Void> response = notificationController.markAllAsRead(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            verify(markAllNotificationsAsReadUseCase).execute(1);
        }
    }

    @Nested
    @DisplayName("getUnreadCount")
    class GetUnreadCount {

        @Test
        @DisplayName("未読通知数を取得できる")
        void returnsUnreadCount() {
            when(getUnreadNotificationCountUseCase.execute(1)).thenReturn(3L);

            ResponseEntity<Long> response = notificationController.getUnreadCount(mockJwt);

            assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
            assertThat(response.getBody()).isEqualTo(3L);
        }
    }
}
