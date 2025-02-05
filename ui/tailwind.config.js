const colors = require('tailwindcss/colors')

module.exports = {
  purge: {
    content: [
      './pages/**/*.{js,ts,jsx,tsx}',
      './src/**/*.{js,ts,jsx,tsx}',
      './node_modules/@mergestat/blocks/**/*.{js,ts,jsx,tsx}',
    ],
    options: {
      safelist: [/w-\d\/\d/, 'md_w-1/2', /^w-/, /^h-/],
    },
  },
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {
      colors: {
        gray: {
          150: '#EBECEF',
          250: '#DDDFE4',
          350: 'B9C0CB',
          450: '#727C8D'
        },
      },
      fontFamily: {
        sans: ['"Inter UI"', 'ui-sans-serif', 'system-ui'],
        mono: ['Monaco', 'ui-monospace', 'SFMono-Regular'],
      },
      gridTemplateRows: {
        'content-layout': 'auto 1fr',
      },
      minHeight: {
        40: '10rem',
      },
      minWidth: {
        sm: '24rem',
      },
      width: {
        1.5: '6px',
        3.5: '14px',
        4.5: '18px',
        100: '25rem',
        150: '37.5rem',
        200: '50rem',
      },
      height: {
        1.5: '6px',
        3.5: '14px',
        4.5: '18px',
        84: '21rem',
      },
      margin: {
        0.5: '2px',
      },
      lineHeight: {
        20: '5rem',
      }
    },
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      gray: colors.coolGray,
      green: colors.emerald,
      lime: colors.lime,
      red: colors.red,
      blue: colors.sky,
      yellow: colors.amber,
      white: colors.white,
    },
  },
  variants: {
    extend: {
      backgroundColor: ['active'],
      backgroundImage: ['active'],
      boxShadow: ['active'],
      gradientColorStops: ['active'],
    },
  },
  plugins: [require('@tailwindcss/line-clamp')],
  separator: '_',
}
