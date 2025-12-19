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
      className="w-full border-2 border-gray-200 rounded-lg py-3 px-4 flex items-center justify-center space-x-3 hover:bg-gradient-to-r hover:from-primary-50 hover:to-secondary-50 hover:border-primary-300 transition-all duration-200 transform hover:scale-105 shadow-sm hover:shadow-md mb-3"
    >
      <img src={providerIcons[provider]} alt={provider} className="w-5 h-5" />
      <span className="text-sm font-semibold text-gray-700">
        {providerLabels[provider]}
      </span>
    </button>
  );
}
