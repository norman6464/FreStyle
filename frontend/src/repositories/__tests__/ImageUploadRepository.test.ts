import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/shared/api/axios', () => ({
  default: { post: vi.fn() },
}));
vi.mock('axios', () => ({
  default: { put: vi.fn() },
}));

import apiClient from '@/shared/api/axios';
import axios from 'axios';
import ImageUploadRepository from '../ImageUploadRepository';

const mockedPost = vi.mocked(apiClient.post);
const mockedPut = vi.mocked(axios.put);

describe('ImageUploadRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('presign → S3 PUT → publicUrl を返す', async () => {
    mockedPost.mockResolvedValue({
      data: { url: 'https://s3/put?sig', key: 'notes/7/1.bin', publicUrl: 'https://cdn/notes/7/1.bin', expiresIn: 600 },
    });
    mockedPut.mockResolvedValue({});

    const file = new File(['x'], 'diagram.png', { type: 'image/png' });
    const url = await ImageUploadRepository.upload(file);

    expect(mockedPost).toHaveBeenCalledWith('/api/v2/notes/images/upload-url', {
      contentType: 'image/png',
    });
    expect(mockedPut).toHaveBeenCalledWith('https://s3/put?sig', file, {
      headers: { 'Content-Type': 'image/png' },
    });
    expect(url).toBe('https://cdn/notes/7/1.bin');
  });
});
