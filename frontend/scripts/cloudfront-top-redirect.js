// トップ(/)にログイン済みの目印があればダッシュボードへ転送する。(FRESTYLE-231)
//
// トップの HTML には SEO 用に紹介ページが焼き込まれており、ブラウザは JS 起動前に
// それを描画する。そのためアプリ内の判定では「一瞬紹介ページが映る」のを防げない。
// 目印(fs_signed_in)を持たない訪問者と検索エンジンのクローラは素通しするので、
// 紹介ページは従来どおり 200 で配信され SEO に影響しない。
function handler(event) {
  var request = event.request;

  if (request.uri !== '/' && request.uri !== '/index.html') {
    return request;
  }

  var cookies = request.cookies || {};
  var hint = cookies['fs_signed_in'];
  if (!hint || hint.value !== '1') {
    return request;
  }

  return {
    statusCode: 302,
    statusDescription: 'Found',
    headers: {
      'location': { value: '/dashboard' },
      // 目印の有無で応答が変わるため、共有キャッシュに載せない。
      'cache-control': { value: 'no-store' }
    }
  };
}
