/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        'trade-green': '#00c853',
        'trade-red': '#ff1744',
        'trade-bg': '#0d1117',
        'trade-card': '#161b22',
        'trade-border': '#30363d',
      },
    },
  },
  plugins: [],
}
