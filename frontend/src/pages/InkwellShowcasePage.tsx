import { useState } from 'react';
import {
  InkwellButton,
  InkwellTextField,
  InkwellCard,
  InkwellCardContent,
  InkwellCardActions,
  InkwellCheckbox,
  InkwellSwitch,
} from '../components/inkwell';

/**
 * inkwell プリミティブの見た目確認用カタログ（開発・レビュー用）。
 * アプリ本体のテーマとは独立した触感的コンポーネント群を一覧する。
 */
export default function InkwellShowcasePage() {
  const [checked, setChecked] = useState(true);
  const [on, setOn] = useState(true);

  return (
    <div className="min-h-screen bg-[#f5f5f5] font-roboto text-inkwell-text-primary">
      <div className="mx-auto max-w-4xl px-6 py-10 space-y-10">
        <header>
          <h1 className="text-3xl font-medium">inkwell UI カタログ</h1>
          <p className="mt-1 text-inkwell-text-secondary">押下波紋・標高シャドウ・浮き上がるラベルを Tailwind だけで実装した触感的プリミティブ。</p>
        </header>

        <Section title="Button — 塗り (contained)">
          <InkwellButton>Primary</InkwellButton>
          <InkwellButton color="secondary">Secondary</InkwellButton>
          <InkwellButton color="error">Error</InkwellButton>
          <InkwellButton disabled>Disabled</InkwellButton>
        </Section>

        <Section title="Button — 枠線 (outlined) / 文字のみ (text)">
          <InkwellButton variant="outlined">Outlined</InkwellButton>
          <InkwellButton variant="outlined" color="secondary">Outlined</InkwellButton>
          <InkwellButton variant="text">Text</InkwellButton>
          <InkwellButton variant="text" color="error">Text</InkwellButton>
        </Section>

        <Section title="Button — サイズ">
          <InkwellButton size="small">Small</InkwellButton>
          <InkwellButton size="medium">Medium</InkwellButton>
          <InkwellButton size="large">Large</InkwellButton>
        </Section>

        <Section title="TextField — 枠線 + 浮き上がるラベル">
          <div className="flex flex-wrap gap-6">
            <InkwellTextField label="お名前" />
            <InkwellTextField label="メール" helperText="社内アドレスを入力" defaultValue="taro@example.jp" />
            <InkwellTextField label="パスワード" type="password" error helperText="8 文字以上にしてください" />
            <InkwellTextField label="無効" disabled />
          </div>
        </Section>

        <Section title="Card — 標高">
          <div className="flex flex-wrap gap-4">
            {([0, 1, 2, 4, 8] as const).map((e) => (
              <InkwellCard key={e} elevation={e} className="w-44">
                <InkwellCardContent>
                  <p className="text-sm font-medium">elevation {e}</p>
                  <p className="mt-1 text-xs text-inkwell-text-secondary">数字が大きいほど浮いて見える</p>
                </InkwellCardContent>
                <InkwellCardActions className="justify-end">
                  <InkwellButton variant="text" size="small">
                    詳細
                  </InkwellButton>
                </InkwellCardActions>
              </InkwellCard>
            ))}
          </div>
        </Section>

        <Section title="Checkbox / Switch">
          <div className="flex flex-wrap items-center gap-6">
            <InkwellCheckbox label="同意する" checked={checked} onChange={(e) => setChecked(e.target.checked)} />
            <InkwellCheckbox label="未チェック" checked={false} readOnly />
            <InkwellCheckbox label="無効" disabled />
            <InkwellSwitch label="通知" checked={on} onChange={(e) => setOn(e.target.checked)} />
            <InkwellSwitch label="無効" disabled />
          </div>
        </Section>
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h2 className="mb-3 text-xs font-semibold uppercase tracking-widest text-inkwell-text-secondary">{title}</h2>
      <div className="flex flex-wrap items-center gap-3">{children}</div>
    </section>
  );
}
