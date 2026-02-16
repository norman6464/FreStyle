package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.repository.NotificationRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetUnreadNotificationCountUseCase {

    private final NotificationRepository notificationRepository;

    @Transactional(readOnly = true)
    public long execute(Integer userId) {
        return notificationRepository.countByUserIdAndIsReadFalse(userId);
    }
}
