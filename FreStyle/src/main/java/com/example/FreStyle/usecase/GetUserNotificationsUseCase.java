package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.NotificationDto;
import com.example.FreStyle.entity.Notification;
import com.example.FreStyle.repository.NotificationRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetUserNotificationsUseCase {

    private final NotificationRepository notificationRepository;

    @Transactional(readOnly = true)
    public List<NotificationDto> execute(Integer userId) {
        List<Notification> notifications = notificationRepository.findByUserIdOrderByCreatedAtDesc(userId);
        return notifications.stream()
                .map(this::toDto)
                .collect(Collectors.toList());
    }

    private NotificationDto toDto(Notification notification) {
        return new NotificationDto(
                notification.getId(),
                notification.getType(),
                notification.getTitle(),
                notification.getMessage(),
                notification.getIsRead(),
                notification.getRelatedId(),
                notification.getCreatedAt().toString()
        );
    }
}
