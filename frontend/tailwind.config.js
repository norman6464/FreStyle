/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#EDEDFC',
          100: '#DBDCF7',
          200: '#B8BAF0',
          300: '#9497E8',
          400: '#7578D9',
          500: '#5B5FC7',
          600: '#4B4EB5',
          700: '#3E41A3',
          800: '#333591',
          900: '#2A2C7A',
        },
      },
      borderRadius: {
        xl: '1rem',
        '2xl': '1.5rem',
      },
      animation: {
        'fade-in': 'fadeIn 0.15s ease-in',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
      },
    },
  },
  plugins: [],
};
