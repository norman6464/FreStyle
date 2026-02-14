package com.example.FreStyle.usecase;

import com.example.FreStyle.dto.NoteDto;
import com.example.FreStyle.repository.NoteRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class GetNotesByUserIdUseCase {

    private final NoteRepository noteRepository;

    public List<NoteDto> execute(Integer userId) {
        return noteRepository.findByUserId(userId);
    }
}
