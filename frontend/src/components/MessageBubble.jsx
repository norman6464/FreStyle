// src/components/MessageBubble.jsx
export default function MessageBubble({ isSender, type = 'text', content }) {
  const baseStyle =
    'max-w-[70%] px-4 py-2 rounded-lg text-sm whitespace-pre-wrap';
  const alignment = isSender
    ? 'self-end bg-blue-500 text-white rounded-br-none'
    : 'self-start bg-white text-gray-800 rounded-bl-none';

  return (
    <div className={`flex ${isSender ? 'justify-end' : 'justify-start'} my-1`}>
      <div className={`${baseStyle} ${alignment}`}>
        {type === 'text' && <p>{content}</p>}
        {type === 'image' && (
          <img src={content} alt="画像" className="max-w-full rounded-md" />
        )}
        {type === 'bot' && <p className="italic text-gray-500">{content}</p>}
      </div>
    </div>
  );
}
