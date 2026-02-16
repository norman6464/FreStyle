package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.SessionNoteDto;
import com.example.FreStyle.repository.SessionNoteRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetSessionNoteUseCase {

    private final SessionNoteRepository sessionNoteRepository;

    @Transactional(readOnly = true)
    public SessionNoteDto execute(Integer userId, Integer sessionId) {
        return sessionNoteRepository.findByUserIdAndSessionId(userId, sessionId)
                .map(note -> new SessionNoteDto(
                        note.getSessionId(),
                        note.getNote(),
                        note.getUpdatedAt().toString()))
                .orElse(null);
    }
}
