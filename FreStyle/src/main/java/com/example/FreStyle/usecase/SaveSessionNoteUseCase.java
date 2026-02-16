package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.SessionNote;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.SessionNoteRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class SaveSessionNoteUseCase {

    private final SessionNoteRepository sessionNoteRepository;

    @Transactional
    public void execute(User user, Integer sessionId, String note) {
        SessionNote sessionNote = sessionNoteRepository.findByUserIdAndSessionId(user.getId(), sessionId)
                .orElseGet(() -> {
                    SessionNote newNote = new SessionNote();
                    newNote.setUser(user);
                    newNote.setSessionId(sessionId);
                    return newNote;
                });
        sessionNote.setNote(note);
        sessionNoteRepository.save(sessionNote);
    }
}
