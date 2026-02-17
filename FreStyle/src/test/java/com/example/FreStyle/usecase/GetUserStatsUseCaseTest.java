package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.UserStatsDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.FriendshipRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetUserStatsUseCase テスト")
class GetUserStatsUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private FriendshipRepository friendshipRepository;

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @InjectMocks
    private GetUserStatsUseCase getUserStatsUseCase;

    @Test
    @DisplayName("ユーザー統計を正しく取得する")
    void execute_returnsCorrectStats() {
        Integer userId = 1;

        AiChatSession normalSession = new AiChatSession();
        normalSession.setSessionType("normal");
        AiChatSession practiceSession1 = new AiChatSession();
        practiceSession1.setSessionType("practice");
        AiChatSession practiceSession2 = new AiChatSession();
        practiceSession2.setSessionType("practice");

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of(normalSession, practiceSession1, practiceSession2));
        when(friendshipRepository.countByFollowingId(userId)).thenReturn(10L);
        when(friendshipRepository.countByFollowerId(userId)).thenReturn(5L);

        CommunicationScore score1 = new CommunicationScore();
        score1.setScore(80);
        CommunicationScore score2 = new CommunicationScore();
        score2.setScore(60);
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of(score1, score2));

        UserStatsDto result = getUserStatsUseCase.execute(userId);

        assertThat(result.totalSessions()).isEqualTo(3);
        assertThat(result.practiceSessionCount()).isEqualTo(2);
        assertThat(result.followerCount()).isEqualTo(10);
        assertThat(result.followingCount()).isEqualTo(5);
        assertThat(result.averageScore()).isEqualTo(70.0);
    }

    @Test
    @DisplayName("セッションが0件の場合は全て0を返す")
    void execute_noSessions_returnsZeros() {
        Integer userId = 2;

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of());
        when(friendshipRepository.countByFollowingId(userId)).thenReturn(0L);
        when(friendshipRepository.countByFollowerId(userId)).thenReturn(0L);
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of());

        UserStatsDto result = getUserStatsUseCase.execute(userId);

        assertThat(result.totalSessions()).isZero();
        assertThat(result.practiceSessionCount()).isZero();
        assertThat(result.followerCount()).isZero();
        assertThat(result.followingCount()).isZero();
        assertThat(result.averageScore()).isEqualTo(0.0);
    }

    @Test
    @DisplayName("normalセッションのみの場合practiceSessionCountは0")
    void execute_onlyNormalSessions() {
        Integer userId = 3;

        AiChatSession session = new AiChatSession();
        session.setSessionType("normal");

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of(session));
        when(friendshipRepository.countByFollowingId(userId)).thenReturn(0L);
        when(friendshipRepository.countByFollowerId(userId)).thenReturn(0L);

        CommunicationScore score = new CommunicationScore();
        score.setScore(90);
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of(score));

        UserStatsDto result = getUserStatsUseCase.execute(userId);

        assertThat(result.totalSessions()).isEqualTo(1);
        assertThat(result.practiceSessionCount()).isZero();
        assertThat(result.averageScore()).isEqualTo(90.0);
    }

    @Test
    @DisplayName("スコアが複数ある場合の平均値が正しい")
    void execute_multipleScores_correctAverage() {
        Integer userId = 4;

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of());
        when(friendshipRepository.countByFollowingId(userId)).thenReturn(0L);
        when(friendshipRepository.countByFollowerId(userId)).thenReturn(0L);

        CommunicationScore s1 = new CommunicationScore();
        s1.setScore(100);
        CommunicationScore s2 = new CommunicationScore();
        s2.setScore(50);
        CommunicationScore s3 = new CommunicationScore();
        s3.setScore(70);
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of(s1, s2, s3));

        UserStatsDto result = getUserStatsUseCase.execute(userId);

        assertThat(result.averageScore()).isCloseTo(73.33, org.assertj.core.data.Offset.offset(0.01));
    }

    @Test
    @DisplayName("sessionTypeがnullの場合はpracticeにカウントされない")
    void execute_nullSessionType_notCountedAsPractice() {
        Integer userId = 5;

        AiChatSession session = new AiChatSession();
        session.setSessionType(null);

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of(session));
        when(friendshipRepository.countByFollowingId(userId)).thenReturn(0L);
        when(friendshipRepository.countByFollowerId(userId)).thenReturn(0L);
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId))
                .thenReturn(List.of());

        UserStatsDto result = getUserStatsUseCase.execute(userId);

        assertThat(result.totalSessions()).isEqualTo(1);
        assertThat(result.practiceSessionCount()).isZero();
    }
}
