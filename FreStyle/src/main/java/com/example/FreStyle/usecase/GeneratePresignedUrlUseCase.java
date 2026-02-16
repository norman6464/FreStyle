package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;

import com.example.FreStyle.dto.PresignedUrlResponse;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.NoteImageService;
import com.example.FreStyle.service.UserIdentityService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GeneratePresignedUrlUseCase {

    private final NoteImageService noteImageService;
    private final UserIdentityService userIdentityService;

    public PresignedUrlResponse execute(String sub, String noteId, String fileName, String contentType) {
        User user = userIdentityService.findUserBySub(sub);
        return noteImageService.generatePresignedUrl(user.getId(), noteId, fileName, contentType);
    }
}
