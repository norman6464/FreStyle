import MessageBubble from './MessageBubble';

interface MessageBubbleAiProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: number;
  onDelete?: (id: number) => void;
  onCopy?: ((id: number, content: string) => void) | null;
  isCopied?: boolean;
  isDeleted?: boolean;
}

export default function MessageBubbleAi(props: MessageBubbleAiProps) {
  return <MessageBubble {...props} />;
}
