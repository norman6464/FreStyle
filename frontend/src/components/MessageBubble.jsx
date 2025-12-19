// isSenderではAIとのチャットでは人間かAIか判断して
// ChatPageでは自分かそれ以外の人かを判断する

export default function MessageBubble({ isSender, type = 'text', content }) {
  const baseStyle =
    'max-w-[70%] px-4 py-3 rounded-2xl text-sm whitespace-pre-wrap break-words shadow-md';
  const alignment = isSender
    ? 'self-end bg-gradient-primary text-white rounded-br-none'
    : 'self-start bg-gray-100 text-gray-800 rounded-bl-none border-l-4 border-primary-400';

  return (
    <div
      className={`flex ${
        isSender ? 'justify-end' : 'justify-start'
      } my-3 animate-fade-in`}
    >
      <div className={`${baseStyle} ${alignment}`}>
        {type === 'text' && <p>{content}</p>}
        {type === 'image' && (
          <img
            src={content}
            alt="画像"
            className="max-w-full rounded-lg shadow-md"
          />
        )}
        {type === 'bot' && <p className="italic opacity-80">{content}</p>}
      </div>
    </div>
  );
}
