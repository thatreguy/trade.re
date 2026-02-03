import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Trade.re - Radically Transparent Trading',
  description: 'A trading game where all statistics are open',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="min-h-screen bg-trade-bg">
          <nav className="border-b border-trade-border bg-trade-card">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="flex justify-between h-16">
                <div className="flex items-center">
                  <span className="text-xl font-bold text-white">Trade.re</span>
                  <span className="ml-2 text-xs bg-purple-600 px-2 py-1 rounded">R.index</span>
                </div>
                <div className="flex items-center space-x-4">
                  <a href="/" className="text-gray-300 hover:text-white px-3 py-2">Trade</a>
                  <a href="/positions" className="text-gray-300 hover:text-white px-3 py-2">All Positions</a>
                  <a href="/leaderboard" className="text-gray-300 hover:text-white px-3 py-2">Leaderboard</a>
                  <a href="/liquidations" className="text-gray-300 hover:text-white px-3 py-2">Liquidations</a>
                </div>
              </div>
            </div>
          </nav>
          <main>{children}</main>
        </div>
      </body>
    </html>
  )
}
