export default function SNSSignInButton({ provider, onClick }) {
  // window.location.href を使うと、SPAのルーティングを飛び越えて、完全にページがリロードされる。

  const providerIcons = {
    google: 'https://developers.google.com/identity/images/g-logo.png',
    facebook:
      'https://upload.wikimedia.org/wikipedia/commons/0/05/Facebook_Logo_%282019%29.png',
    x: 'https://cdn.cms%E2%80%91twdigitalassets.com/content/dam/about-twitter/x/brand-toolkit/x-white-logo.png',
  };

  const providerLabels = {
    google: 'Googleでログイン',
    facebook: 'Facebookでログイン',
    x: 'Xでログイン',
  };

  return (
    <button
      onClick={onClick}
      className="w-full border border-gray-300 rounded-lg py-2.5 px-4 flex items-center justify-center space-x-3 hover:bg-gray-50 transition-colors duration-150 mb-3"
    >
      <img src={providerIcons[provider]} alt={provider} className="w-5 h-5" />
      <span className="text-sm font-medium text-gray-700">
        {providerLabels[provider]}
      </span>
    </button>
  );
}
