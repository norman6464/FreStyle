package com.example.FreStyle.usecase;

import com.example.FreStyle.repository.NoteRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class UpdateNoteUseCase {

    private final NoteRepository noteRepository;

    public void execute(Integer userId, String noteId, String title, String content, Boolean isPinned) {
        noteRepository.update(userId, noteId, title, content, isPinned);
    }
}
