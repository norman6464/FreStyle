import apiClient from '../lib/axios';
import { DAILY_GOALS, PROFILE } from '../constants/apiRoutes';

export interface ProfileStats {
  totalSessions: number;
  averageScore: number;
  currentStreak: number;
  longestStreak: number;
  totalAchievedDays: number;
  followerCount: number;
  followingCount: number;
}

const ProfileStatsRepository = {
  async fetchStats(): Promise<ProfileStats> {
    const [statsRes, streakRes] = await Promise.all([
      apiClient.get(PROFILE.meStats),
      apiClient.get(DAILY_GOALS.streak),
    ]);
    return {
      totalSessions: statsRes.data.totalSessions ?? 0,
      averageScore: statsRes.data.averageScore ?? 0,
      followerCount: statsRes.data.followerCount ?? 0,
      followingCount: statsRes.data.followingCount ?? 0,
      currentStreak: streakRes.data.currentStreak ?? 0,
      longestStreak: streakRes.data.longestStreak ?? 0,
      totalAchievedDays: streakRes.data.totalAchievedDays ?? 0,
    };
  },
};

export default ProfileStatsRepository;
