package com.example.FreStyle.usecase;

import com.example.FreStyle.dto.NoteDto;
import com.example.FreStyle.repository.NoteRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class CreateNoteUseCase {

    private final NoteRepository noteRepository;

    public NoteDto execute(Integer userId, String title) {
        return noteRepository.save(userId, title);
    }
}
