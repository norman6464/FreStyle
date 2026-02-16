import { useState, useCallback } from 'react';
import ProfileRepository from '../repositories/ProfileRepository';

const ALLOWED_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp'];

export function useProfileImageUpload() {
  const [uploading, setUploading] = useState(false);

  const upload = useCallback(async (file: File): Promise<string | null> => {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return null;
    }

    setUploading(true);
    try {
      const { uploadUrl, imageUrl } = await ProfileRepository.getImagePresignedUrl(file.name, file.type);
      await ProfileRepository.uploadToS3(uploadUrl, file);
      return imageUrl;
    } catch (error) {
      console.error('プロフィール画像のアップロードに失敗しました:', error);
      return null;
    } finally {
      setUploading(false);
    }
  }, []);

  return { upload, uploading };
}
