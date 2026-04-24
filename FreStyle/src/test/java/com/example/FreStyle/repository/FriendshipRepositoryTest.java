package com.example.FreStyle.repository;

import static org.assertj.core.api.Assertions.assertThat;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase;
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest;
import org.springframework.boot.test.autoconfigure.orm.jpa.TestEntityManager;

import com.example.FreStyle.entity.Friendship;
import com.example.FreStyle.entity.User;

/**
 * FriendshipRepository の統合テスト。
 * H2 インメモリDB上で実際にJPAクエリを走らせ、カスタムクエリメソッドの挙動を担保する。
 * Infrastructure層の代表的なテストパターンとしてRepository単位の@DataJpaTestを採用。
 *
 * Note: 本番の schema.sql は MariaDB 固有構文を含むため、テストでは無効化して
 * Hibernate による create-drop で H2 上にスキーマを作る方針。
 */
@DataJpaTest(properties = {
        // 本番 schema.sql (MariaDB構文) を無効化し、H2 上にエンティティから create する
        "spring.sql.init.mode=never",
        "spring.jpa.hibernate.ddl-auto=create-drop",
        // H2 dialect を明示（application.properties の MariaDBDialect を上書き）
        "spring.jpa.properties.hibernate.dialect=org.hibernate.dialect.H2Dialect"
})
@AutoConfigureTestDatabase
@DisplayName("FriendshipRepository 統合テスト")
class FriendshipRepositoryTest {

    @Autowired
    private FriendshipRepository friendshipRepository;

    @Autowired
    private TestEntityManager entityManager;

    private User persistUser(String name, String email) {
        User user = new User();
        user.setName(name);
        user.setEmail(email);
        user.setIsActive(true);
        return entityManager.persistAndFlush(user);
    }

    private Friendship persistFriendship(User follower, User following, LocalDateTime createdAt) {
        Friendship friendship = new Friendship();
        friendship.setFollower(follower);
        friendship.setFollowing(following);
        friendship.setCreatedAt(createdAt);
        return entityManager.persistAndFlush(friendship);
    }

    @Nested
    @DisplayName("findByFollowerIdOrderByCreatedAtDesc")
    class FindByFollowerId {

        @Test
        @DisplayName("指定ユーザーがフォローしている関係を作成日の新しい順で返す")
        void returnsDescendingByCreatedAt() {
            User me = persistUser("me", "me@example.com");
            User a = persistUser("a", "a@example.com");
            User b = persistUser("b", "b@example.com");

            persistFriendship(me, a, LocalDateTime.of(2026, 1, 1, 12, 0));
            persistFriendship(me, b, LocalDateTime.of(2026, 2, 1, 12, 0));

            List<Friendship> result = friendshipRepository.findByFollowerIdOrderByCreatedAtDesc(me.getId());

            assertThat(result).hasSize(2);
            assertThat(result.get(0).getFollowing().getName()).isEqualTo("b");
            assertThat(result.get(1).getFollowing().getName()).isEqualTo("a");
        }

        @Test
        @DisplayName("フォローしていない場合は空のリストを返す")
        void returnsEmptyWhenNoFollowing() {
            User me = persistUser("me", "me@example.com");
            List<Friendship> result = friendshipRepository.findByFollowerIdOrderByCreatedAtDesc(me.getId());
            assertThat(result).isEmpty();
        }
    }

    @Nested
    @DisplayName("findByFollowerIdAndFollowingId / existsByFollowerIdAndFollowingId")
    class FindOne {

        @Test
        @DisplayName("既存のフォロー関係を1件取得できる")
        void findsExistingFriendship() {
            User me = persistUser("me", "me@example.com");
            User target = persistUser("target", "target@example.com");
            persistFriendship(me, target, LocalDateTime.now());

            Optional<Friendship> result = friendshipRepository.findByFollowerIdAndFollowingId(me.getId(), target.getId());

            assertThat(result).isPresent();
            assertThat(result.get().getFollowing().getId()).isEqualTo(target.getId());
        }

        @Test
        @DisplayName("存在しないフォロー関係はOptional.emptyを返す")
        void returnsEmptyWhenNotExists() {
            User me = persistUser("me", "me@example.com");
            User target = persistUser("target", "target@example.com");

            Optional<Friendship> result = friendshipRepository.findByFollowerIdAndFollowingId(me.getId(), target.getId());

            assertThat(result).isEmpty();
        }

        @Test
        @DisplayName("existsByFollowerIdAndFollowingId は boolean を返す")
        void existsReturnsBoolean() {
            User me = persistUser("me", "me@example.com");
            User target = persistUser("target", "target@example.com");
            persistFriendship(me, target, LocalDateTime.now());

            assertThat(friendshipRepository.existsByFollowerIdAndFollowingId(me.getId(), target.getId())).isTrue();
            assertThat(friendshipRepository.existsByFollowerIdAndFollowingId(target.getId(), me.getId())).isFalse();
        }
    }

    @Nested
    @DisplayName("countByFollowerId / countByFollowingId")
    class CountTests {

        @Test
        @DisplayName("フォロー数を正しく返す")
        void countsFollowing() {
            User me = persistUser("me", "me@example.com");
            User a = persistUser("a", "a@example.com");
            User b = persistUser("b", "b@example.com");
            persistFriendship(me, a, LocalDateTime.now());
            persistFriendship(me, b, LocalDateTime.now());

            assertThat(friendshipRepository.countByFollowerId(me.getId())).isEqualTo(2);
        }

        @Test
        @DisplayName("フォロワー数を正しく返す")
        void countsFollowers() {
            User target = persistUser("target", "target@example.com");
            User a = persistUser("a", "a@example.com");
            User b = persistUser("b", "b@example.com");
            User c = persistUser("c", "c@example.com");
            persistFriendship(a, target, LocalDateTime.now());
            persistFriendship(b, target, LocalDateTime.now());
            persistFriendship(c, target, LocalDateTime.now());

            assertThat(friendshipRepository.countByFollowingId(target.getId())).isEqualTo(3);
        }
    }

    @Nested
    @DisplayName("findMutualFollowerIds / findMutualFollowingIds（相互フォロー判定）")
    class MutualTests {

        @Test
        @DisplayName("targetIdをフォローしているuserIdのみを返す")
        void findsMutualFollowerIds() {
            User me = persistUser("me", "me@example.com");
            User target = persistUser("target", "target@example.com");
            User other = persistUser("other", "other@example.com");

            // me と other が target をフォロー
            persistFriendship(me, target, LocalDateTime.now());
            persistFriendship(other, target, LocalDateTime.now());
            // target は me だけをフォロー
            persistFriendship(target, me, LocalDateTime.now());

            List<Integer> result = friendshipRepository.findMutualFollowerIds(
                    List.of(me.getId(), other.getId()), target.getId());

            assertThat(result).containsExactlyInAnyOrder(me.getId(), other.getId());
        }

        @Test
        @DisplayName("targetIdにフォローされているuserIdのみを返す")
        void findsMutualFollowingIds() {
            User me = persistUser("me", "me@example.com");
            User target = persistUser("target", "target@example.com");
            User other = persistUser("other", "other@example.com");

            // target は me と other をフォロー
            persistFriendship(target, me, LocalDateTime.now());
            persistFriendship(target, other, LocalDateTime.now());

            List<Integer> result = friendshipRepository.findMutualFollowingIds(
                    List.of(me.getId(), other.getId()), target.getId());

            assertThat(result).containsExactlyInAnyOrder(me.getId(), other.getId());
        }

        @Test
        @DisplayName("候補リストが空のとき、空のリストを返す")
        void returnsEmptyWhenNoCandidates() {
            User target = persistUser("target", "target@example.com");
            List<Integer> result = friendshipRepository.findMutualFollowerIds(List.of(), target.getId());
            assertThat(result).isEmpty();
        }
    }
}
