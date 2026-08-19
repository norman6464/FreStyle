import { Link } from 'react-router-dom';

/** セクション間で繰り返す小型 CTA(白ピル + 申請ボタン)。反復導線で申請への距離を縮める。 */
export default function CtaStrip({ text }: { text: string }) {
  return (
    <div className="flex justify-center bg-white pb-14">
      <span className="inline-flex items-center gap-4 rounded-full border border-slate-200 bg-white py-2.5 pl-6 pr-3 text-[13.5px] font-bold text-slate-600 shadow-[0_10px_30px_-18px_rgba(23,36,92,0.35)]">
        {text}
        <Link
          to="/company-application"
          className="rounded-full bg-brand-600 px-5 py-2 text-[13px] font-bold text-white transition hover:bg-brand-700"
        >
          導入・利用申請
        </Link>
      </span>
    </div>
  );
}
