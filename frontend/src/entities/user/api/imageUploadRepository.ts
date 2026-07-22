import apiClient from '@/shared/api/axios';
import axios from 'axios';
import { IMAGES } from '@/shared/config/apiRoutes';

interface UploadUrlResponse {
  url: string; // S3 への PUT 用署名付き URL
  key: string;
  publicUrl: string; // アップロード後に参照する配信 URL
  expiresIn: number;
}

/**
 * ImageUploadRepository — 画像を S3 にアップロードして公開 URL を返す（ノート / 教材で共有）。
 *
 * フロー: presign 発行(current user 名義) → S3 へ直接 PUT → 配信 URL(publicUrl) を返す。
 * userId は送らず backend が context の current user で発行する（IDOR 対策）。
 */
const ImageUploadRepository = {
  async upload(file: File): Promise<string> {
    const { data } = await apiClient.post<UploadUrlResponse>(IMAGES.uploadUrl, {
      contentType: file.type || 'image/png',
    });
    await axios.put(data.url, file, {
      headers: { 'Content-Type': file.type || 'image/png' },
    });
    return data.publicUrl;
  },
};

export default ImageUploadRepository;
