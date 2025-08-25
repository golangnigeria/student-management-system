/** @type {import('tailwindcss').Config} */
module.exports = {
    content: [
    "./web/templates/**/*.{html,templ,go}",
    "./cmd/web/**/*.{go,html,templ}",
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}

