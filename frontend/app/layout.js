import './globals.css'

export const metadata = { title: 'Text Vault', description: 'A simple bucket-backed text file vault' }

export default function RootLayout({ children }) {
  return <html lang="en"><body>{children}</body></html>
}
