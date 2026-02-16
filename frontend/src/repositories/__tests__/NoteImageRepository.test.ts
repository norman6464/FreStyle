import { describe, it, expect, vi, beforeEach } from 'vitest';
import NoteImageRepository from '../NoteImageRepository';
import apiClient from '../../lib/axios';
import axios from 'axios';

vi.mock('../../lib/axios', () => ({
  default: {
    post: vi.fn(),
  },
}));

vi.mock('axios', () => ({
  default: {
    put: vi.fn(),
  },
}));

describe('NoteImageRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getPresignedUrl', () => {
    it('presigned URLを取得する', async () => {
      const mockResponse = {
        data: {
          uploadUrl: 'https://s3.example.com/upload',
          imageUrl: 'https://cdn.example.com/notes/1/note1/abc_image.png',
        },
      };
      vi.mocked(apiClient.post).mockResolvedValue(mockResponse);

      const result = await NoteImageRepository.getPresignedUrl('note1', 'image.png', 'image/png');

      expect(apiClient.post).toHaveBeenCalledWith('/api/notes/note1/images/presigned-url', {
        fileName: 'image.png',
        contentType: 'image/png',
      });
      expect(result.uploadUrl).toBe('https://s3.example.com/upload');
      expect(result.imageUrl).toBe('https://cdn.example.com/notes/1/note1/abc_image.png');
    });
  });

  describe('uploadToS3', () => {
    it('S3にファイルをアップロードする', async () => {
      vi.mocked(axios.put).mockResolvedValue({ status: 200 });
      const file = new File(['test'], 'image.png', { type: 'image/png' });

      await NoteImageRepository.uploadToS3('https://s3.example.com/upload', file);

      expect(axios.put).toHaveBeenCalledWith('https://s3.example.com/upload', file, {
        headers: { 'Content-Type': 'image/png' },
      });
    });

    it('S3アップロード失敗時にエラーが伝搬する', async () => {
      vi.mocked(axios.put).mockRejectedValue(new Error('Network Error'));
      const file = new File(['test'], 'image.png', { type: 'image/png' });

      await expect(NoteImageRepository.uploadToS3('https://s3.example.com/upload', file))
        .rejects.toThrow('Network Error');
    });
  });

  describe('getPresignedUrl エラー', () => {
    it('API失敗時にエラーが伝搬する', async () => {
      vi.mocked(apiClient.post).mockRejectedValue(new Error('Server Error'));

      await expect(NoteImageRepository.getPresignedUrl('note1', 'image.png', 'image/png'))
        .rejects.toThrow('Server Error');
    });
  });
});
