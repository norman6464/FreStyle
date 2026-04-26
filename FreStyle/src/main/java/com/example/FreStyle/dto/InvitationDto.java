package com.example.FreStyle.dto;

import java.time.LocalDateTime;

public record InvitationDto(
        Long id,
        Long companyId,
        String email,
        String role,
        Long invitedBy,
        LocalDateTime expiresAt,
        LocalDateTime acceptedAt,
        Long acceptedUserId,
        LocalDateTime createdAt
) {}
