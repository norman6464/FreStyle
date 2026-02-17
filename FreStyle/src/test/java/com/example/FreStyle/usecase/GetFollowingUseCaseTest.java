package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.time.LocalDateTime;
import java.util.List;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.Spy;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.FriendshipDto;
import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.FriendshipMapper;
import com.example.FreStyle.repository.FriendshipRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetFollowingUseCase テスト")
class GetFollowingUseCaseTest {

    @Mock
    private FriendshipRepository friendshipRepository;

    @Spy
    private FriendshipMapper friendshipMapper;

    @InjectMocks
    private GetFollowingUseCase getFollowingUseCase;

    private User user;
    private User target1;
    private User target2;

    @BeforeEach
    void setUp() {
        user = new User();
        user.setId(1);
        user.setName("自分");
        target1 = new User();
        target1.setId(2);
        target1.setName("ユーザーB");
        target2 = new User();
        target2.setId(3);
        target2.setName("ユーザーC");
    }

    @Test
    @DisplayName("フォロー中一覧を取得できる")
    void execute_returnsFollowingList() {
        Friendship f1 = new Friendship(1, user, target1, LocalDateTime.now());
        Friendship f2 = new Friendship(2, user, target2, LocalDateTime.now());

        when(friendshipRepository.findByFollowerIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(f1, f2));
        when(friendshipRepository.findMutualFollowerIds(List.of(2, 3), 1))
                .thenReturn(List.of(2));

        List<FriendshipDto> result = getFollowingUseCase.execute(1);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).userId()).isEqualTo(2);
        assertThat(result.get(0).mutual()).isTrue();
        assertThat(result.get(1).userId()).isEqualTo(3);
        assertThat(result.get(1).mutual()).isFalse();
    }

    @Test
    @DisplayName("フォローしていない場合は空リストを返す")
    void execute_returnsEmptyList() {
        when(friendshipRepository.findByFollowerIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());

        List<FriendshipDto> result = getFollowingUseCase.execute(1);

        assertThat(result).isEmpty();
    }
}
