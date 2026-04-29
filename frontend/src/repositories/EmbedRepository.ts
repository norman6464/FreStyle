import apiClient from '../lib/axios';
import { EMBEDS } from '../constants/apiRoutes';

/**
 * 外部 URL の OGP / oEmbed メタを Go backend のプロキシ経由で取得する。
 * バックエンドが SSRF 対策・キャッシュを担うため、フロントは薄い wrapper でよい。
 */
export interface EmbedCardDto {
  url: string;
  title?: string;
  description?: string;
  imageUrl?: string;
  siteName?: string;
  /** "ogp" / "youtube" / "github" / ... バックエンドが解決した戦略 */
  provider?: string;
}

export const EmbedRepository = {
  async resolve(url: string): Promise<EmbedCardDto> {
    const res = await apiClient.get<EmbedCardDto>(EMBEDS.oembed, { params: { url } });
    return res.data;
  },
};
