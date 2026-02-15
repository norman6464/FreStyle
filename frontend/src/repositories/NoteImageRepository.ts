import apiClient from '../lib/axios';
import axios from 'axios';

interface PresignedUrlResponse {
  uploadUrl: string;
  imageUrl: string;
}

const NoteImageRepository = {
  async getPresignedUrl(noteId: string, fileName: string, contentType: string): Promise<PresignedUrlResponse> {
    const res = await apiClient.post(`/api/notes/${noteId}/images/presigned-url`, {
      fileName,
      contentType,
    });
    return res.data;
  },

  async uploadToS3(uploadUrl: string, file: File): Promise<void> {
    await axios.put(uploadUrl, file, {
      headers: { 'Content-Type': file.type },
    });
  },
};

export default NoteImageRepository;
