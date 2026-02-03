package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thatreguy/trade.re/internal/api"
	"github.com/thatreguy/trade.re/internal/domain"
	"github.com/thatreguy/trade.re/internal/engine"
	"github.com/thatreguy/trade.re/internal/ws"
)

func main() {
	// Initialize matching engine
	eng := engine.NewMatchingEngine()

	// Register R.index - the only tradeable instrument
	eng.RegisterInstrument("R.index")

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Wire up trade broadcasts
	eng.OnTrade(func(trade *domain.Trade) {
		hub.BroadcastTrade(trade)
		log.Printf("Trade: %s %s @ %s (buyer: %s, seller: %s)",
			trade.Size.String(),
			trade.Instrument,
			trade.Price.String(),
			trade.BuyerID.String()[:8],
			trade.SellerID.String()[:8],
		)
	})

	eng.OnOrderUpdate(func(order *domain.Order) {
		hub.Broadcast(ws.Message{
			Type: ws.TypeOrder,
			Data: order,
		})
	})

	// Create API server
	server := api.NewServer(eng, hub)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// Register routes
	server.RegisterRoutes(r)

	// Get port from env or default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("=================================")
	log.Printf("  Trade.re Server Starting")
	log.Printf("  Port: %s", port)
	log.Printf("  Instrument: R.index")
	log.Printf("=================================")
	log.Printf("")
	log.Printf("Endpoints:")
	log.Printf("  GET  /health")
	log.Printf("  GET  /ws (WebSocket)")
	log.Printf("  GET  /api/v1/auth/register")
	log.Printf("  GET  /api/v1/auth/login")
	log.Printf("  GET  /api/v1/traders")
	log.Printf("  GET  /api/v1/traders/{id}")
	log.Printf("  GET  /api/v1/traders/{id}/positions")
	log.Printf("  GET  /api/v1/market/orderbook")
	log.Printf("  GET  /api/v1/market/positions")
	log.Printf("  GET  /api/v1/market/trades")
	log.Printf("  GET  /api/v1/market/stats")
	log.Printf("  POST /api/v1/orders")
	log.Printf("  DELETE /api/v1/orders/{id}")
	log.Printf("")

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
