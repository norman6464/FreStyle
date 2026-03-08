package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.UserStatsDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.FriendshipRepository;
import com.example.FreStyle.service.UserService;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetUserStatsUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final FriendshipRepository friendshipRepository;
    private final CommunicationScoreRepository communicationScoreRepository;
    private final UserService userService;

    @Transactional(readOnly = true)
    public UserStatsDto execute(Integer userId) {
        // ユーザー存在チェック（存在しない場合はRuntimeException）
        userService.findUserById(userId);
        List<AiChatSession> sessions = aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId);
        long totalSessions = sessions.size();
        long practiceSessionCount = sessions.stream()
                .filter(s -> "practice".equals(s.getSessionType()))
                .count();

        long followerCount = friendshipRepository.countByFollowingId(userId);
        long followingCount = friendshipRepository.countByFollowerId(userId);

        List<CommunicationScore> scores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);
        double averageScore = scores.stream()
                .mapToInt(CommunicationScore::getScore)
                .average()
                .orElse(0.0);

        return new UserStatsDto(totalSessions, practiceSessionCount, followerCount, followingCount, averageScore);
    }
}
