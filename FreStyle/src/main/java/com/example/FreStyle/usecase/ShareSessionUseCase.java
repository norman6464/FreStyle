package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.SharedSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.SharedSession;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.form.ShareSessionForm;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.SharedSessionRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class ShareSessionUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final SharedSessionRepository sharedSessionRepository;

    @Transactional
    public SharedSessionDto execute(Integer userId, ShareSessionForm form) {
        AiChatSession session = aiChatSessionRepository.findByIdAndUserId(form.sessionId(), userId)
                .orElseThrow(() -> new ResourceNotFoundException(
                    "セッションが見つからないか、アクセス権限がありません: sessionId=" + form.sessionId() + ", userId=" + userId
                ));

        SharedSession sharedSession = new SharedSession();
        sharedSession.setSession(session);
        sharedSession.setUser(session.getUser());
        sharedSession.setDescription(form.description());
        sharedSession.setIsPublic(true);

        SharedSession saved = sharedSessionRepository.save(sharedSession);

        return new SharedSessionDto(
                saved.getId(),
                saved.getSession().getId(),
                saved.getSession().getTitle(),
                saved.getUser().getId(),
                saved.getUser().getName(),
                saved.getUser().getIconUrl(),
                saved.getDescription(),
                saved.getCreatedAt() != null ? saved.getCreatedAt().toLocalDateTime().toString() : null
        );
    }
}
