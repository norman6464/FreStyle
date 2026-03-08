package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.SharedSessionDto;
import com.example.FreStyle.entity.SharedSession;
import com.example.FreStyle.repository.SharedSessionRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetPublicSessionsUseCase {

    private final SharedSessionRepository sharedSessionRepository;

    public List<SharedSessionDto> execute() {
        List<SharedSession> sessions = sharedSessionRepository.findPublicSessions();
        return sessions.stream()
                .map(ss -> new SharedSessionDto(
                        ss.getId(),
                        ss.getSession().getId(),
                        ss.getSession().getTitle(),
                        ss.getUser().getId(),
                        ss.getUser().getName(),
                        ss.getUser().getIconUrl(),
                        ss.getDescription(),
                        ss.getCreatedAt() != null ? ss.getCreatedAt().toLocalDateTime().toString() : null
                ))
                .toList();
    }
}
