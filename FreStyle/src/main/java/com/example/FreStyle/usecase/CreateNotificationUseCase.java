package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.Notification;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.NotificationRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class CreateNotificationUseCase {

    private final NotificationRepository notificationRepository;

    @Transactional
    public void execute(User user, String type, String title, String message, Integer relatedId) {
        Notification notification = new Notification();
        notification.setUser(user);
        notification.setType(type);
        notification.setTitle(title);
        notification.setMessage(message);
        notification.setRelatedId(relatedId);
        notificationRepository.save(notification);
    }
}
