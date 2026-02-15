import { ChatBubbleLeftRightIcon, EnvelopeIcon, AcademicCapIcon, PresentationChartBarIcon } from '@heroicons/react/24/outline';
import { TEMPLATE_CATEGORIES, CONVERSATION_TEMPLATES } from '../constants/conversationTemplates';
import type { TemplateCategory } from '../constants/conversationTemplates';

interface ConversationTemplatesProps {
  onSelect: (prompt: string) => void;
}

const CATEGORY_ICONS: Record<TemplateCategory['iconName'], React.ComponentType<React.SVGProps<SVGSVGElement>>> = {
  'envelope': EnvelopeIcon,
  'academic-cap': AcademicCapIcon,
  'chat-bubble': ChatBubbleLeftRightIcon,
  'presentation': PresentationChartBarIcon,
};

export default function ConversationTemplates({ onSelect }: ConversationTemplatesProps) {
  return (
    <div className="max-w-2xl mx-auto w-full space-y-6">
      <p className="text-sm font-medium text-[var(--color-text-secondary)] text-center">
        テンプレートから始める
      </p>

      <div className="flex flex-wrap justify-center gap-2 mb-2">
        {TEMPLATE_CATEGORIES.map(({ name, iconName }) => {
          const Icon = CATEGORY_ICONS[iconName];
          return (
            <span key={name} className="inline-flex items-center gap-1 text-xs text-[var(--color-text-tertiary)] bg-surface-2 px-2.5 py-1 rounded-full">
              <Icon className="w-3.5 h-3.5" />
              {name}
            </span>
          );
        })}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {CONVERSATION_TEMPLATES.map((template) => (
          <button
            key={template.title}
            onClick={() => onSelect(template.prompt)}
            className="text-left p-3 rounded-lg border border-surface-3 hover:border-[var(--color-text-muted)] hover:bg-surface-2 transition-colors group"
          >
            <p className="text-xs font-medium text-[var(--color-text-primary)] group-hover:text-primary-400 transition-colors">
              {template.title}
            </p>
            <p className="text-[10px] text-[var(--color-text-tertiary)] mt-0.5 line-clamp-2">
              {template.prompt}
            </p>
          </button>
        ))}
      </div>
    </div>
  );
}
