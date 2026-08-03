import { describe, it, expect, vi, beforeEach } from 'vitest';
import aiChatRepository from '../aiChatRepository';
import apiClient from '@/shared/api/axios';
import axios from 'axios';

vi.mock('@/shared/api/axios', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
    patch: vi.fn(),
  },
}));

vi.mock('axios', () => ({
  default: {
    put: vi.fn(),
  },
}));

/**
 * 添付ファイルのアップロード（FRESTYLE-22）。
 *
 * 以前は画面（MessageInput）が直接 fetch していた。通信は repository 層に置く。
 */
describe('aiChatRepository の添付アップロード', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('presigned URL へファイル本体を PUT する', async () => {
    vi.mocked(axios.put).mockResolvedValue({});
    const file = new File(['dummy'], 'shot.png', { type: 'image/png' });

    await aiChatRepository.uploadAttachment('https://s3.example.com/upload?sig=x', file);

    expect(axios.put).toHaveBeenCalledWith('https://s3.example.com/upload?sig=x', file, {
      headers: { 'Content-Type': 'image/png' },
    });
  });

  // 送信先は S3 であって自分の API ではない。apiClient を使うと認証 Cookie が付き、
  // 401 で /login へ飛ばす割り込みも動いてしまう。
  it('自分の API 用クライアントは使わない', async () => {
    vi.mocked(axios.put).mockResolvedValue({});
    const file = new File(['dummy'], 'shot.png', { type: 'image/png' });

    await aiChatRepository.uploadAttachment('https://s3.example.com/upload', file);

    expect(apiClient.put).not.toHaveBeenCalled();
    expect(apiClient.post).not.toHaveBeenCalled();
  });

  it('アップロードの失敗は呼び出し側へ伝える', async () => {
    vi.mocked(axios.put).mockRejectedValue(new Error('network error'));
    const file = new File(['dummy'], 'shot.png', { type: 'image/png' });

    await expect(
      aiChatRepository.uploadAttachment('https://s3.example.com/upload', file),
    ).rejects.toThrow('network error');
  });

  it('presigned URL の発行は自分の API へ問い合わせる', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({
      data: { uploadUrl: 'https://s3.example.com/u', key: 'ai-chat/1/x.png', expiresIn: 600 },
    });

    const res = await aiChatRepository.issueAttachmentUploadUrl({
      filename: 'shot.png',
      contentType: 'image/png',
      sizeBytes: 123,
    });

    expect(apiClient.post).toHaveBeenCalled();
    expect(res.key).toBe('ai-chat/1/x.png');
  });
});
