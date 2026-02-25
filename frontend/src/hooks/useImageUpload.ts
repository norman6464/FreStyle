import { useCallback, useEffect, useRef, useState } from 'react';
import type { Editor } from '@tiptap/core';
import NoteImageRepository from '../repositories/NoteImageRepository';

const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp', 'image/svg+xml'];

export function useImageUpload(noteId: string | null, editor: Editor | null) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);

  const uploadAndInsert = useCallback(async (file: File) => {
    if (!noteId || !editor) return;
    if (!ALLOWED_TYPES.includes(file.type)) return;
    setUploadError(null);

    try {
      const { uploadUrl, imageUrl } = await NoteImageRepository.getPresignedUrl(
        noteId, file.name, file.type
      );
      await NoteImageRepository.uploadToS3(uploadUrl, file);
      editor.chain().focus().setImage({ src: imageUrl, alt: file.name }).run();
    } catch {
      setUploadError('画像アップロードに失敗しました');
    }
  }, [noteId, editor]);

  const handleFilesFromInput = useCallback((files: FileList | null) => {
    if (!files) return;
    Array.from(files).forEach(file => uploadAndInsert(file));
  }, [uploadAndInsert]);

  const openFileDialog = useCallback(() => {
    if (!fileInputRef.current) {
      const input = document.createElement('input');
      input.type = 'file';
      input.accept = 'image/*';
      input.style.display = 'none';
      input.addEventListener('change', () => {
        handleFilesFromInput(input.files);
        input.value = '';
      });
      document.body.appendChild(input);
      fileInputRef.current = input;
    }
    fileInputRef.current.click();
  }, [handleFilesFromInput]);

  useEffect(() => {
    return () => {
      if (fileInputRef.current) {
        fileInputRef.current.remove();
        fileInputRef.current = null;
      }
    };
  }, []);

  const handleDrop = useCallback((event: React.DragEvent) => {
    const files = event.dataTransfer?.files;
    if (!files || files.length === 0) return;

    const imageFiles = Array.from(files).filter(f => ALLOWED_TYPES.includes(f.type));
    if (imageFiles.length === 0) return;

    event.preventDefault();
    imageFiles.forEach(file => uploadAndInsert(file));
  }, [uploadAndInsert]);

  const handlePaste = useCallback((event: React.ClipboardEvent) => {
    const items = event.clipboardData?.items;
    if (!items) return;

    for (const item of Array.from(items)) {
      if (item.type.startsWith('image/')) {
        const file = item.getAsFile();
        if (file) {
          event.preventDefault();
          uploadAndInsert(file);
          return;
        }
      }
    }
  }, [uploadAndInsert]);

  return { uploadAndInsert, openFileDialog, handleDrop, handlePaste, uploadError };
}
