import apiClient from '../lib/axios';
import { COMPANY_SETTINGS } from '../constants/apiRoutes';

/** 会社設定（trainee への AI エージェント機能の有効/無効）。 */
export interface CompanySettings {
  aiChatEnabledForTrainees: boolean;
}

/**
 * 会社設定 API の薄いラッパ。company_admin / super_admin のみが利用する。
 * axios の直接利用はこのレイヤに集約する。
 */
const CompanySettingsRepository = {
  /** 自社の設定を取得する。 */
  async get(): Promise<CompanySettings> {
    const res = await apiClient.get<CompanySettings>(COMPANY_SETTINGS.base);
    return res.data;
  },

  /** 自社の AI 有効化フラグを更新する。 */
  async update(aiChatEnabledForTrainees: boolean): Promise<CompanySettings> {
    const res = await apiClient.put<CompanySettings>(COMPANY_SETTINGS.base, {
      aiChatEnabledForTrainees,
    });
    return res.data;
  },
};

export default CompanySettingsRepository;
