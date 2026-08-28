import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
// アプリ全体の書体（Noto Sans JP）。CSP が font-src 'self' なので CDN ではなく
// 自前配信（unicode-range 分割済みで、使う文字の分だけダウンロードされる）。
// 重みはアプリで使う 4 つだけ（400 標準 / 500 medium / 600 semibold / 700 bold）。
import '@fontsource/noto-sans-jp/400.css';
import '@fontsource/noto-sans-jp/500.css';
import '@fontsource/noto-sans-jp/600.css';
import '@fontsource/noto-sans-jp/700.css';
import './styles/index.css';
import { Provider } from 'react-redux';
import { store } from '@/app/store';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <BrowserRouter>
    <Provider store={store}>
      <App />
    </Provider>
  </BrowserRouter>
);
