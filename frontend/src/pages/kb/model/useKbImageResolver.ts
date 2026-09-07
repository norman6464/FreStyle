import { useCallback, useRef } from 'react';
import { KbRepository } from '@/entities/kb';

/** ナレッジの画像 src が「解決の要る S3 key」であることを示す接頭辞。 */
const KB_KEY_PREFIX = 'kb/';

/**
 * キャッシュの持ち時間。backend の presigned URL の期限（10 分）より短く固定しておく
 * ことで、「キャッシュは有効期限内のつもりだったが、表示した瞬間には S3 側で
 * 切れていた」という取りこぼしを避ける。
 */
const CACHE_TTL_MS = 9 * 60 * 1000;

interface CacheEntry {
  url: string;
  expiresAt: number;
}

/**
 * useKbImageResolver は、本文・カバー画像が doc に持つ S3 の key
 * （"kb/<workspaceId>/<pageId>/<epochNs>.bin"）を、表示に使える期限付き URL へ解決する
 * resolveImageSrc を返す。
 *
 * - "kb/" で始まらない src（http(s)://・data: 等）はそのまま返す。解決できない/
 *   解決する必要が無いものには触れない（既存のテスト・story・データとの互換のため）。
 * - 同じ key への解決要求は Map でキャッシュする。ページを開いている間ずっと保持する
 *   ref なので、同じ画像が複数回描画されても S3 の署名 URL 発行を毎回叩かない。
 * - 解決に失敗したら**例外を投げる**（呼び出し側の ImageView が失敗表示に倒す）。
 */
export function useKbImageResolver(workspaceSlug: string | undefined, pageId: string | undefined) {
  // ページを開いている間だけ生きるキャッシュ。ref なので再描画のたびに作り直さない。
  const cache = useRef(new Map<string, CacheEntry>());

  const resolveImageSrc = useCallback(
    async (src: string): Promise<string> => {
      if (!src.startsWith(KB_KEY_PREFIX)) return src;

      const cached = cache.current.get(src);
      if (cached && cached.expiresAt > Date.now()) {
        return cached.url;
      }

      if (!workspaceSlug || !pageId) {
        // ページの所属がまだ解決していない（通常は起こらない — resolveImageSrc は
        // ページが開けてから配線される）。解決しようがないので例外を投げる。
        throw new Error('kb page is not resolved yet');
      }

      const { url } = await KbRepository.issuePageImageDownloadURL(workspaceSlug, pageId, src);
      cache.current.set(src, { url, expiresAt: Date.now() + CACHE_TTL_MS });
      return url;
    },
    [workspaceSlug, pageId],
  );

  return { resolveImageSrc };
}
