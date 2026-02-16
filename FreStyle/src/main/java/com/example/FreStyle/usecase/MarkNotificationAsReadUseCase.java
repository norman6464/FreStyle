package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.Notification;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.NotificationRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class MarkNotificationAsReadUseCase {

    private final NotificationRepository notificationRepository;

    @Transactional
    public void execute(Integer userId, Integer notificationId) {
        Notification notification = notificationRepository.findById(notificationId)
                .orElseThrow(() -> new ResourceNotFoundException("通知が見つかりません。"));

        if (!notification.getUser().getId().equals(userId)) {
            throw new ResourceNotFoundException("この通知にアクセスする権限がありません。");
        }

        notification.setIsRead(true);
        notificationRepository.save(notification);
    }
}
