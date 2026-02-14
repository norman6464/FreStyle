package com.example.FreStyle.usecase;

import com.example.FreStyle.repository.NoteRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class DeleteNoteUseCase {

    private final NoteRepository noteRepository;

    public void execute(Integer userId, String noteId) {
        noteRepository.delete(userId, noteId);
    }
}
