import { useCallback, useEffect, useRef, useState } from 'react';
import { EditorContent } from '@tiptap/react';
import 'tippy.js/dist/tippy.css';
import { executeCommand } from '../extensions/SlashCommandExtension';
import { useBlockEditor } from '../hooks/useBlockEditor';
import { useImageUpload } from '../hooks/useImageUpload';
import { useLinkEditor } from '../hooks/useLinkEditor';
import BlockInserterButton from './BlockInserterButton';
import EmojiPicker from './EmojiPicker';
import LinkBubbleMenu from './LinkBubbleMenu';
import SearchReplaceBar from './SearchReplaceBar';
import SelectionToolbar from './SelectionToolbar';
import type { SlashCommand } from '../constants/slashCommands';
import { UI_TIMINGS } from '../constants/uiTimings';

interface BlockEditorProps {
  content: string;
  onChange: (jsonString: string) => void;
  noteId: string | null;
}

export default function BlockEditor({ content, onChange, noteId }: BlockEditorProps) {
  const { editor } = useBlockEditor({ content, onChange });

  const containerRef = useRef<HTMLDivElement>(null);
  const [inserterVisible, setInserterVisible] = useState(false);
  const [inserterTop, setInserterTop] = useState(0);
  const inserterMenuOpen = useRef(false);
  const [searchOpen, setSearchOpen] = useState(false);
  const [emojiPickerOpen, setEmojiPickerOpen] = useState(false);
  const [youtubeInputOpen, setYoutubeInputOpen] = useState(false);
  const [youtubeUrl, setYoutubeUrl] = useState('');
  const youtubeInputRef = useRef<HTMLInputElement>(null);

  const { openFileDialog, handleDrop, handlePaste } = useImageUpload(noteId, editor);
  const { linkBubble, handleEditorClick, handleEditLink, handleRemoveLink } = useLinkEditor(editor, containerRef);

  // スラッシュコマンドから画像アップロード・絵文字ピッカー・YouTube入力を呼べるようにする
  const openEmojiPicker = useCallback(() => setEmojiPickerOpen(true), []);
  const openYoutubeInput = useCallback(() => {
    setYoutubeInputOpen(true);
    setYoutubeUrl('');
    setTimeout(() => youtubeInputRef.current?.focus(), 0);
  }, []);

  useEffect(() => {
    if (!editor) return;
    editor.storage.slashCommand.onImageUpload = openFileDialog;
    editor.storage.slashCommand.onEmojiPicker = openEmojiPicker;
    editor.storage.slashCommand.onYoutubeUrl = openYoutubeInput;
    return () => {
      editor.storage.slashCommand.onImageUpload = null;
      editor.storage.slashCommand.onEmojiPicker = null;
      editor.storage.slashCommand.onYoutubeUrl = null;
    };
  }, [editor, openFileDialog, openEmojiPicker, openYoutubeInput]);

  // Ctrl+F / Cmd+F で検索バーを開く
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'f') {
        e.preventDefault();
        setSearchOpen(prev => !prev);
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  const lastMoveTime = useRef(0);
  const hideTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearHideTimeout = useCallback(() => {
    if (hideTimeout.current) {
      clearTimeout(hideTimeout.current);
      hideTimeout.current = null;
    }
  }, []);

  const scheduleHide = useCallback(() => {
    clearHideTimeout();
    hideTimeout.current = setTimeout(() => {
      if (!inserterMenuOpen.current) {
        setInserterVisible(false);
      }
    }, UI_TIMINGS.INSERTER_HIDE_DELAY);
  }, [clearHideTimeout]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const now = Date.now();
    if (now - lastMoveTime.current < UI_TIMINGS.MOUSE_MOVE_THROTTLE) return;
    lastMoveTime.current = now;

    if (!containerRef.current) return;
    const editorEl = containerRef.current.querySelector('.ProseMirror');
    if (!editorEl) return;

    const target = e.target as HTMLElement;

    // +ボタン上にマウスがある場合は表示状態を維持
    if (target.closest('[data-block-inserter]')) {
      clearHideTimeout();
      return;
    }

    const block = target.closest('p, h1, h2, h3, ul, ol, li, blockquote');
    if (!block || !editorEl.contains(block)) {
      scheduleHide();
      return;
    }

    clearHideTimeout();
    const containerRect = containerRef.current.getBoundingClientRect();
    const blockRect = block.getBoundingClientRect();
    setInserterTop(blockRect.top - containerRect.top);
    setInserterVisible(true);
  }, [clearHideTimeout, scheduleHide]);

  const handleMouseLeave = useCallback(() => {
    if (!inserterMenuOpen.current) {
      scheduleHide();
    }
  }, [scheduleHide]);

  const handleInserterMenuOpenChange = useCallback((open: boolean) => {
    inserterMenuOpen.current = open;
    if (!open) {
      setInserterVisible(false);
    }
  }, []);

  const handleEmojiSelect = useCallback((emoji: string) => {
    if (!editor) return;
    editor.chain().focus().insertContent(emoji).run();
  }, [editor]);

  const submitYoutubeUrl = useCallback(() => {
    if (!editor || !youtubeUrl.trim()) {
      setYoutubeInputOpen(false);
      setYoutubeUrl('');
      return;
    }
    const setYoutubeVideo = (editor.commands as { setYoutubeVideo?: (opts: { src: string }) => boolean }).setYoutubeVideo;
    if (typeof setYoutubeVideo === 'function') {
      setYoutubeVideo({ src: youtubeUrl.trim() });
    }
    setYoutubeInputOpen(false);
    setYoutubeUrl('');
  }, [editor, youtubeUrl]);

  const handleCommand = useCallback((command: SlashCommand) => {
    if (!editor) return;
    if (command.action === 'image') {
      openFileDialog();
      return;
    }
    if (command.action === 'emoji') {
      setEmojiPickerOpen(true);
      return;
    }
    if (command.action === 'youtube') {
      openYoutubeInput();
      return;
    }
    editor.chain().focus().run();
    executeCommand(editor, command);
  }, [editor, openFileDialog]);

  return (
    <div
      ref={containerRef}
      className="block-editor flex-1 overflow-y-auto relative pl-10"
      data-testid="block-editor"
      onMouseMove={handleMouseMove}
      onMouseLeave={handleMouseLeave}
      onDrop={handleDrop}
      onPaste={handlePaste}
      onDragOver={(e) => e.preventDefault()}
      onClick={handleEditorClick}
    >
      <SearchReplaceBar editor={editor} isOpen={searchOpen} onClose={() => setSearchOpen(false)} />
      <SelectionToolbar editor={editor} containerRef={containerRef} />
      {linkBubble && (
        <div
          className="absolute z-50"
          style={{ top: linkBubble.top, left: linkBubble.left }}
        >
          <LinkBubbleMenu
            url={linkBubble.url}
            onEdit={handleEditLink}
            onRemove={handleRemoveLink}
          />
        </div>
      )}
      <BlockInserterButton
        visible={inserterVisible}
        top={inserterTop}
        onCommand={handleCommand}
        onMenuOpenChange={handleInserterMenuOpenChange}
      />
      <EditorContent editor={editor} aria-label="ノートの内容" />
      {emojiPickerOpen && (
        <div className="absolute z-50 top-8 left-8">
          <EmojiPicker
            isOpen={emojiPickerOpen}
            onSelect={handleEmojiSelect}
            onClose={() => setEmojiPickerOpen(false)}
          />
        </div>
      )}
      {youtubeInputOpen && (
        <div className="absolute z-50 top-8 left-8 bg-[var(--color-surface-1)] border border-[var(--color-surface-3)] rounded-lg shadow-xl p-3">
          <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-2">YouTubeのURLを入力</p>
          <div className="flex items-center gap-2">
            <input
              ref={youtubeInputRef}
              type="url"
              value={youtubeUrl}
              onChange={(e) => setYoutubeUrl(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') { e.preventDefault(); submitYoutubeUrl(); }
                if (e.key === 'Escape') { setYoutubeInputOpen(false); setYoutubeUrl(''); }
              }}
              placeholder="https://www.youtube.com/watch?v=..."
              className="text-xs bg-[var(--color-surface-2)] border border-[var(--color-surface-3)] rounded px-2 py-1.5 text-[var(--color-text-primary)] placeholder:text-[var(--color-text-faint)] w-64 outline-none focus:border-primary-400"
              aria-label="YouTube URL"
            />
            <button
              type="button"
              className="text-xs px-3 py-1.5 rounded bg-primary-500 text-white hover:bg-primary-600 transition-colors"
              onClick={submitYoutubeUrl}
            >
              追加
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
