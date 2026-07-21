import MessageBubble, { type MessageAttachmentView } from './MessageBubble';

interface MessageBubbleAiProps {
  isSender: boolean;
  type?: 'text' | 'image' | 'bot';
  content: string;
  id: string;
  attachments?: MessageAttachmentView[];
  onDelete?: (id: string) => void;
  onCopy?: ((id: string, content: string) => void) | null;
  isCopied?: boolean;
  isDeleted?: boolean;
}

export default function MessageBubbleAi(props: MessageBubbleAiProps) {
  return <MessageBubble {...props} />;
}
