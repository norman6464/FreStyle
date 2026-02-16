package com.example.FreStyle.controller;

import java.util.List;

import org.springframework.http.ResponseEntity;
import org.springframework.security.core.annotation.AuthenticationPrincipal;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import com.example.FreStyle.dto.NotificationDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.usecase.GetUnreadNotificationCountUseCase;
import com.example.FreStyle.usecase.GetUserNotificationsUseCase;
import com.example.FreStyle.usecase.MarkAllNotificationsAsReadUseCase;
import com.example.FreStyle.usecase.MarkNotificationAsReadUseCase;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

@RestController
@RequiredArgsConstructor
@RequestMapping("/api/notifications")
@Slf4j
public class NotificationController {

    private final GetUserNotificationsUseCase getUserNotificationsUseCase;
    private final MarkNotificationAsReadUseCase markNotificationAsReadUseCase;
    private final MarkAllNotificationsAsReadUseCase markAllNotificationsAsReadUseCase;
    private final GetUnreadNotificationCountUseCase getUnreadNotificationCountUseCase;
    private final UserIdentityService userIdentityService;

    @GetMapping
    public ResponseEntity<List<NotificationDto>> getNotifications(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        List<NotificationDto> notifications = getUserNotificationsUseCase.execute(user.getId());
        return ResponseEntity.ok(notifications);
    }

    @PutMapping("/{notificationId}/read")
    public ResponseEntity<Void> markAsRead(
            @AuthenticationPrincipal Jwt jwt,
            @PathVariable Integer notificationId) {
        User user = resolveUser(jwt);
        markNotificationAsReadUseCase.execute(user.getId(), notificationId);
        return ResponseEntity.ok().build();
    }

    @PutMapping("/read-all")
    public ResponseEntity<Void> markAllAsRead(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        markAllNotificationsAsReadUseCase.execute(user.getId());
        return ResponseEntity.ok().build();
    }

    @GetMapping("/unread-count")
    public ResponseEntity<Long> getUnreadCount(@AuthenticationPrincipal Jwt jwt) {
        User user = resolveUser(jwt);
        long count = getUnreadNotificationCountUseCase.execute(user.getId());
        return ResponseEntity.ok(count);
    }

    private User resolveUser(Jwt jwt) {
        return userIdentityService.findUserBySub(jwt.getSubject());
    }
}
