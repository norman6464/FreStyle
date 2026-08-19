/** セクション見出し(eyebrow + タイトル + リード)。dark は濃紺セクション用に配色だけ差し替える。 */
export default function SectionHead({
  eyebrow,
  title,
  lede,
  dark,
}: {
  eyebrow: string;
  title: string;
  lede?: string;
  dark?: boolean;
}) {
  return (
    <div className="text-center">
      <p className={`text-xs font-bold tracking-[0.16em] ${dark ? 'text-[#9db8f5]' : 'text-brand-700'}`}>
        {eyebrow}
      </p>
      <h2
        className={`mt-3 text-[1.75rem] font-extrabold leading-snug tracking-tight sm:text-[2rem] ${dark ? 'text-white' : ''}`}
      >
        {title}
      </h2>
      {lede && <p className={`mt-3 ${dark ? 'text-[#c7d5f7]' : 'text-slate-600'}`}>{lede}</p>}
    </div>
  );
}
