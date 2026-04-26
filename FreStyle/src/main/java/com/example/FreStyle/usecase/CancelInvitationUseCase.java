package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.Invitation;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.InvitationRepository;

import lombok.RequiredArgsConstructor;

/**
 * 自社の招待をキャンセルするユースケース（DB から物理削除）。
 * 自社所有でない場合は ResourceNotFoundException を投げる。
 */
@Service
@RequiredArgsConstructor
public class CancelInvitationUseCase {

    private final InvitationRepository repository;

    @Transactional
    public void execute(Long companyId, Long invitationId) {
        Invitation inv = repository.findById(invitationId)
                .orElseThrow(() -> new ResourceNotFoundException("招待 id=" + invitationId + " が見つかりません"));
        if (!inv.getCompanyId().equals(companyId)) {
            throw new ResourceNotFoundException("招待 id=" + invitationId + " は自社所有ではありません");
        }
        repository.delete(inv);
    }
}
