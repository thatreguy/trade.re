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
              <div className="flex justify-between h-14 sm:h-16">
                <div className="flex items-center">
                  <a href="/" className="text-lg sm:text-xl font-bold text-white">Trade.re</a>
                  <span className="ml-2 text-xs bg-purple-600 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded">R.index</span>
                </div>
                {/* Desktop nav */}
                <div className="hidden sm:flex items-center space-x-1 md:space-x-4">
                  <a href="/" className="text-gray-300 hover:text-white px-2 md:px-3 py-2 text-sm md:text-base">Trade</a>
                  <a href="/positions" className="text-gray-300 hover:text-white px-2 md:px-3 py-2 text-sm md:text-base">Positions</a>
                  <a href="/leaderboard" className="text-gray-300 hover:text-white px-2 md:px-3 py-2 text-sm md:text-base">Leaderboard</a>
                  <a href="/liquidations" className="text-gray-300 hover:text-white px-2 md:px-3 py-2 text-sm md:text-base">Liquidations</a>
                </div>
              </div>
              {/* Mobile nav */}
              <div className="sm:hidden flex justify-around py-2 border-t border-trade-border -mx-4 px-4">
                <a href="/" className="text-gray-300 hover:text-white text-xs px-2 py-1">Trade</a>
                <a href="/positions" className="text-gray-300 hover:text-white text-xs px-2 py-1">Positions</a>
                <a href="/leaderboard" className="text-gray-300 hover:text-white text-xs px-2 py-1">Leaders</a>
                <a href="/liquidations" className="text-gray-300 hover:text-white text-xs px-2 py-1">Liqs</a>
              </div>
            </div>
          </nav>
          <main>{children}</main>
        </div>
      </body>
    </html>
  )
}
