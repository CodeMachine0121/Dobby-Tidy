/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      colors: {
        primary: {
          DEFAULT: '#2563EB',
          hover: '#1D4ED8',
          light: '#EFF6FF',
        },
        accent: {
          DEFAULT: '#D97706',
          light: '#FEF3C7',
        },
        success: {
          DEFAULT: '#16A34A',
          light: '#F0FDF4',
        },
        destructive: {
          DEFAULT: '#DC2626',
          light: '#FEF2F2',
        },
        muted: '#F1F5FD',
        border: '#E4ECFC',
        surface: '#F8FAFC',
        sidebar: '#0F172A',
        'sidebar-hover': '#1E293B',
        'sidebar-active': '#1E3A8A',
      },
    },
  },
  plugins: [],
}
