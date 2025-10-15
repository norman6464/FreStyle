import { atom } from 'recoil';

// JWTのアクセストークンを管理する
export const accessTokenState = atom({
  key: 'accessTokenState',
  default: null,
});
