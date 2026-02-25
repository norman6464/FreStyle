import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useImageUpload } from '../useImageUpload';
import NoteImageRepository from '../../repositories/NoteImageRepository';

vi.mock('../../repositories/NoteImageRepository', () => ({
  default: {
    getPresignedUrl: vi.fn(),
    uploadToS3: vi.fn(),
  },
}));

function createMockEditor() {
  const run = vi.fn();
  const setImage = vi.fn(() => ({ run }));
  const focus = vi.fn(() => ({ setImage }));
  const chain = vi.fn(() => ({ focus }));
  return { chain, focus, setImage, run } as unknown as ReturnType<typeof createMockEditor> & {
    chain: ReturnType<typeof vi.fn>;
    focus: ReturnType<typeof vi.fn>;
    setImage: ReturnType<typeof vi.fn>;
    run: ReturnType<typeof vi.fn>;
  };
}

describe('useImageUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('画像ファイルをアップロードしてエディタに挿入する', async () => {
    const editor = createMockEditor();
    vi.mocked(NoteImageRepository.getPresignedUrl).mockResolvedValue({
      uploadUrl: 'https://s3.example.com/upload',
      imageUrl: 'https://cdn.example.com/image.png',
    });
    vi.mocked(NoteImageRepository.uploadToS3).mockResolvedValue(undefined);

    const { result } = renderHook(() => useImageUpload('note1', editor as any));

    const file = new File(['test'], 'photo.png', { type: 'image/png' });
    await act(async () => {
      await result.current.uploadAndInsert(file);
    });

    expect(NoteImageRepository.getPresignedUrl).toHaveBeenCalledWith('note1', 'photo.png', 'image/png');
    expect(NoteImageRepository.uploadToS3).toHaveBeenCalledWith('https://s3.example.com/upload', file);
    expect(editor.chain).toHaveBeenCalled();
  });

  it('noteIdがnullの場合はアップロードしない', async () => {
    const editor = createMockEditor();
    const { result } = renderHook(() => useImageUpload(null, editor as any));

    const file = new File(['test'], 'photo.png', { type: 'image/png' });
    await act(async () => {
      await result.current.uploadAndInsert(file);
    });

    expect(NoteImageRepository.getPresignedUrl).not.toHaveBeenCalled();
  });

  it('許可されていないファイル形式はアップロードしない', async () => {
    const editor = createMockEditor();
    const { result } = renderHook(() => useImageUpload('note1', editor as any));

    const file = new File(['test'], 'file.pdf', { type: 'application/pdf' });
    await act(async () => {
      await result.current.uploadAndInsert(file);
    });

    expect(NoteImageRepository.getPresignedUrl).not.toHaveBeenCalled();
  });

  it('エラー時にuploadErrorが設定される', async () => {
    const editor = createMockEditor();
    vi.mocked(NoteImageRepository.getPresignedUrl).mockRejectedValue(new Error('network error'));

    const { result } = renderHook(() => useImageUpload('note1', editor as any));

    const file = new File(['test'], 'photo.png', { type: 'image/png' });
    await act(async () => {
      await result.current.uploadAndInsert(file);
    });

    expect(result.current.uploadError).toBe('画像アップロードに失敗しました');
  });

  it('再アップロード成功時にuploadErrorがクリアされる', async () => {
    const editor = createMockEditor();
    vi.mocked(NoteImageRepository.getPresignedUrl)
      .mockRejectedValueOnce(new Error('network error'))
      .mockResolvedValueOnce({
        uploadUrl: 'https://s3.example.com/upload',
        imageUrl: 'https://cdn.example.com/image.png',
      });
    vi.mocked(NoteImageRepository.uploadToS3).mockResolvedValue(undefined);

    const { result } = renderHook(() => useImageUpload('note1', editor as any));

    const file = new File(['test'], 'photo.png', { type: 'image/png' });
    await act(async () => {
      await result.current.uploadAndInsert(file);
    });
    expect(result.current.uploadError).toBe('画像アップロードに失敗しました');

    await act(async () => {
      await result.current.uploadAndInsert(file);
    });
    expect(result.current.uploadError).toBeNull();
  });
});
