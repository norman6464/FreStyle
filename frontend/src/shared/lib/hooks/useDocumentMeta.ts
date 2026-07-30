import { useEffect } from 'react';

/**
 * ページ単位で `<title>` / meta description / canonical / robots を設定する軽量フック。
 *
 * react-helmet を追加せず標準 DOM API で head を upsert する。ビルド時プリレンダー
 * (Playwright スナップショット) が、描画後の head を静的 HTML に取り込めるようにするのが目的。
 * SPA なので次ページの呼び出しで上書きされる想定（未設定項目は前ページの値が残るが、
 * index.html に既定値があるため空にはならない）。
 */
export interface DocumentMeta {
  title?: string;
  description?: string;
  /** 絶対 URL。省略時は現在の origin + pathname を使う。 */
  canonical?: string;
  /** 例: "noindex, nofollow"。認証必須ページで使う。 */
  robots?: string;
}

function upsertMeta(attr: 'name' | 'property', key: string, content: string): void {
  let el = document.head.querySelector<HTMLMetaElement>(`meta[${attr}="${key}"]`);
  if (!el) {
    el = document.createElement('meta');
    el.setAttribute(attr, key);
    document.head.appendChild(el);
  }
  el.setAttribute('content', content);
}

function upsertLink(rel: string, href: string): void {
  let el = document.head.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`);
  if (!el) {
    el = document.createElement('link');
    el.setAttribute('rel', rel);
    document.head.appendChild(el);
  }
  el.setAttribute('href', href);
}

export function useDocumentMeta({ title, description, canonical, robots }: DocumentMeta): void {
  useEffect(() => {
    if (title) {
      document.title = title;
    }
    if (description) {
      upsertMeta('name', 'description', description);
    }
    if (robots) {
      upsertMeta('name', 'robots', robots);
    } else {
      // 未指定時は既存タグを削除する。認証画面(noindex)から公開LPへ SPA 遷移したとき
      // noindex が残留すると公開ページまで noindex になってしまうため。
      document.head.querySelector('meta[name="robots"]')?.remove();
    }
    const href = canonical ?? `${window.location.origin}${window.location.pathname}`;
    upsertLink('canonical', href);
  }, [title, description, canonical, robots]);
}
