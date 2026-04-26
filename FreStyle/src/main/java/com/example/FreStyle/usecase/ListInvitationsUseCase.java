package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.InvitationDto;
import com.example.FreStyle.entity.Invitation;
import com.example.FreStyle.repository.InvitationRepository;

import lombok.RequiredArgsConstructor;

/**
 * 自社の未承諾招待を一覧取得するユースケース。
 */
@Service
@RequiredArgsConstructor
public class ListInvitationsUseCase {

    private final InvitationRepository repository;

    @Transactional(readOnly = true)
    public List<InvitationDto> execute(Long companyId) {
        return repository.findByCompanyIdAndAcceptedAtIsNullOrderByCreatedAtDesc(companyId)
                .stream().map(this::toDto).toList();
    }

    private InvitationDto toDto(Invitation i) {
        return new InvitationDto(
                i.getId(), i.getCompanyId(), i.getEmail(), i.getRole(),
                i.getInvitedBy(), i.getExpiresAt(), i.getAcceptedAt(),
                i.getAcceptedUserId(), i.getCreatedAt()
        );
    }
}
