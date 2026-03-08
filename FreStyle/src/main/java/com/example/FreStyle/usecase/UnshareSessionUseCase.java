package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.SharedSession;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.SharedSessionRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class UnshareSessionUseCase {

    private final SharedSessionRepository sharedSessionRepository;

    @Transactional
    public void execute(Integer userId, Integer sessionId) {
        SharedSession sharedSession = sharedSessionRepository.findBySessionId(sessionId)
                .orElseThrow(() -> new ResourceNotFoundException(
                    "共有セッションが見つかりません: sessionId=" + sessionId
                ));

        if (!sharedSession.getUser().getId().equals(userId)) {
            throw new ResourceNotFoundException(
                "共有セッションへのアクセス権限がありません: sessionId=" + sessionId + ", userId=" + userId
            );
        }

        sharedSessionRepository.delete(sharedSession);
    }
}
