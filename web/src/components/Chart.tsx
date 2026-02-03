'use client'

import { useEffect, useRef } from 'react'
import { createChart, ColorType, IChartApi, ISeriesApi, CandlestickData, Time } from 'lightweight-charts'
import { useMarketStore } from '@/store/market'

export default function Chart() {
  const chartContainerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<IChartApi | null>(null)
  const candleSeriesRef = useRef<ISeriesApi<'Candlestick'> | null>(null)
  const { recentTrades } = useMarketStore()

  useEffect(() => {
    if (!chartContainerRef.current) return

    // Create chart
    const chart = createChart(chartContainerRef.current, {
      layout: {
        background: { type: ColorType.Solid, color: '#1a1a2e' },
        textColor: '#9ca3af',
      },
      grid: {
        vertLines: { color: '#2d2d44' },
        horzLines: { color: '#2d2d44' },
      },
      width: chartContainerRef.current.clientWidth,
      height: 340,
      timeScale: {
        timeVisible: true,
        secondsVisible: false,
        borderColor: '#2d2d44',
      },
      rightPriceScale: {
        borderColor: '#2d2d44',
      },
      crosshair: {
        mode: 1,
        vertLine: {
          color: '#6366f1',
          width: 1,
          style: 2,
        },
        horzLine: {
          color: '#6366f1',
          width: 1,
          style: 2,
        },
      },
    })

    chartRef.current = chart

    // Add candlestick series
    const candleSeries = chart.addCandlestickSeries({
      upColor: '#22c55e',
      downColor: '#ef4444',
      borderUpColor: '#22c55e',
      borderDownColor: '#ef4444',
      wickUpColor: '#22c55e',
      wickDownColor: '#ef4444',
    })

    candleSeriesRef.current = candleSeries

    // Handle resize
    const handleResize = () => {
      if (chartContainerRef.current) {
        chart.applyOptions({ width: chartContainerRef.current.clientWidth })
      }
    }

    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      chart.remove()
    }
  }, [])

  // Update chart data from trades
  useEffect(() => {
    if (!candleSeriesRef.current || recentTrades.length === 0) return

    // Group trades into 1-minute candles
    const candles = new Map<number, { open: number; high: number; low: number; close: number; time: number }>()

    // Sort trades by timestamp (oldest first)
    const sortedTrades = [...recentTrades].sort((a, b) =>
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    )

    sortedTrades.forEach(trade => {
      const price = parseFloat(trade.price) || 0
      if (price === 0) return

      const tradeTime = new Date(trade.timestamp)
      // Round to minute
      const minuteTime = Math.floor(tradeTime.getTime() / 60000) * 60

      if (!candles.has(minuteTime)) {
        candles.set(minuteTime, {
          time: minuteTime,
          open: price,
          high: price,
          low: price,
          close: price,
        })
      } else {
        const candle = candles.get(minuteTime)!
        candle.high = Math.max(candle.high, price)
        candle.low = Math.min(candle.low, price)
        candle.close = price
      }
    })

    // Convert to array and sort
    const candleData: CandlestickData<Time>[] = Array.from(candles.values())
      .sort((a, b) => a.time - b.time)
      .map(c => ({
        time: c.time as Time,
        open: c.open,
        high: c.high,
        low: c.low,
        close: c.close,
      }))

    if (candleData.length > 0) {
      candleSeriesRef.current.setData(candleData)
    }
  }, [recentTrades])

  return (
    <div ref={chartContainerRef} className="w-full h-full" />
  )
}
