package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.RankingDto;
import com.example.FreStyle.service.RankingService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
@Transactional(readOnly = true)
public class GetRankingUseCase {

    private final RankingService rankingService;

    public RankingDto execute(String period, Integer currentUserId) {
        return rankingService.getRanking(period, currentUserId);
    }
}
