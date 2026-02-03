package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/thatreguy/trade.re/internal/domain"
	"github.com/thatreguy/trade.re/internal/engine"
	"github.com/thatreguy/trade.re/internal/ws"
)

// Server holds the API dependencies
type Server struct {
	engine   *engine.MatchingEngine
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

// NewServer creates a new API server
func NewServer(eng *engine.MatchingEngine, hub *ws.Hub) *Server {
	return &Server{
		engine: eng,
		hub:    hub,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
	}
}

// Response helpers
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// RegisterRoutes sets up all API routes
func (s *Server) RegisterRoutes(r chi.Router) {
	// Health check
	r.Get("/health", s.handleHealth)

	// WebSocket endpoint
	r.Get("/ws", s.handleWebSocket)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Traders (public - transparency!)
		r.Route("/traders", func(r chi.Router) {
			r.Get("/", s.handleGetTraders)
			r.Post("/", s.handleCreateTrader)
			r.Get("/{traderID}", s.handleGetTrader)
			r.Get("/{traderID}/positions", s.handleGetTraderPositions)
		})

		// Instruments
		r.Route("/instruments", func(r chi.Router) {
			r.Get("/{symbol}/orderbook", s.handleGetOrderBook)
			r.Get("/{symbol}/positions", s.handleGetPositions)
			r.Get("/{symbol}/oi", s.handleGetOpenInterest)
		})

		// Market (convenience routes for R.index)
		r.Route("/market", func(r chi.Router) {
			r.Get("/orderbook", s.handleGetMarketOrderBook)
			r.Get("/positions", s.handleGetMarketPositions)
			r.Get("/oi", s.handleGetMarketOpenInterest)
			r.Get("/trades", s.handleGetMarketTrades)
			r.Get("/liquidations", s.handleGetMarketLiquidations)
			r.Get("/stats", s.handleGetMarketStats)
		})

		// Auth (simplified for now)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
		})

		// Orders
		r.Route("/orders", func(r chi.Router) {
			r.Post("/", s.handleSubmitOrder)
			r.Delete("/{orderID}", s.handleCancelOrder)
		})
	})
}

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// handleWebSocket upgrades to WebSocket connection
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(s.hub, conn)
	s.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// handleGetTraders returns all traders (public)
func (s *Server) handleGetTraders(w http.ResponseWriter, r *http.Request) {
	traders := s.engine.GetAllTraders()
	respondJSON(w, http.StatusOK, traders)
}

// handleCreateTrader registers a new trader
func (s *Server) handleCreateTrader(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string           `json:"username"`
		Type     domain.TraderType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" {
		respondError(w, http.StatusBadRequest, "username is required")
		return
	}

	if req.Type == "" {
		req.Type = domain.TraderTypeHuman
	}

	trader := &domain.Trader{
		ID:        uuid.New(),
		Username:  req.Username,
		Type:      req.Type,
		CreatedAt: time.Now(),
		TotalPnL:  decimal.Zero,
	}

	s.engine.RegisterTrader(trader)
	respondJSON(w, http.StatusCreated, trader)
}

// handleGetTrader returns a single trader (public)
func (s *Server) handleGetTrader(w http.ResponseWriter, r *http.Request) {
	traderIDStr := chi.URLParam(r, "traderID")
	traderID, err := uuid.Parse(traderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid trader ID")
		return
	}

	trader := s.engine.GetTrader(traderID)
	if trader == nil {
		respondError(w, http.StatusNotFound, "trader not found")
		return
	}

	respondJSON(w, http.StatusOK, trader)
}

// handleGetTraderPositions returns a trader's positions (public - transparency!)
func (s *Server) handleGetTraderPositions(w http.ResponseWriter, r *http.Request) {
	traderIDStr := chi.URLParam(r, "traderID")
	traderID, err := uuid.Parse(traderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid trader ID")
		return
	}

	// Get R.index position
	positions := make([]*domain.Position, 0)
	pos := s.engine.GetPosition(traderID, "R.index")
	if pos != nil && !pos.Size.IsZero() {
		positions = append(positions, pos)
	}

	respondJSON(w, http.StatusOK, positions)
}

// handleGetOrderBook returns the order book (public)
func (s *Server) handleGetOrderBook(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	depthStr := r.URL.Query().Get("depth")
	depth := 20
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 && d <= 100 {
			depth = d
		}
	}

	book, err := s.engine.GetOrderBook(symbol, depth)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, book)
}

// handleGetPositions returns all positions for an instrument (public - transparency!)
func (s *Server) handleGetPositions(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	positions := s.engine.GetAllPositions(symbol)
	respondJSON(w, http.StatusOK, positions)
}

// handleGetOpenInterest returns OI breakdown (the key transparency feature!)
func (s *Server) handleGetOpenInterest(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")
	oi := s.engine.GetOpenInterestBreakdown(symbol)
	respondJSON(w, http.StatusOK, oi)
}

// handleSubmitOrder submits a new order
func (s *Server) handleSubmitOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TraderID   string `json:"trader_id"`
		Instrument string `json:"instrument"`
		Side       string `json:"side"`
		Type       string `json:"type"`
		Price      string `json:"price"`
		Size       string `json:"size"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	traderID, err := uuid.Parse(req.TraderID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid trader_id")
		return
	}

	price, err := decimal.NewFromString(req.Price)
	if err != nil && req.Type == "limit" {
		respondError(w, http.StatusBadRequest, "invalid price")
		return
	}

	size, err := decimal.NewFromString(req.Size)
	if err != nil || size.LessThanOrEqual(decimal.Zero) {
		respondError(w, http.StatusBadRequest, "invalid size")
		return
	}

	order := &domain.Order{
		TraderID:   traderID,
		Instrument: req.Instrument,
		Side:       domain.Side(req.Side),
		Type:       domain.OrderType(req.Type),
		Price:      price,
		Size:       size,
	}

	trades, err := s.engine.SubmitOrder(order)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Broadcast trades via WebSocket
	for _, trade := range trades {
		s.hub.BroadcastTrade(trade)
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"order":  order,
		"trades": trades,
	})
}

// handleCancelOrder cancels an existing order
func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	orderIDStr := chi.URLParam(r, "orderID")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid order ID")
		return
	}

	instrument := r.URL.Query().Get("instrument")
	if instrument == "" {
		respondError(w, http.StatusBadRequest, "instrument is required")
		return
	}

	if err := s.engine.CancelOrder(orderID, instrument); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// Market convenience routes for R.index

func (s *Server) handleGetMarketOrderBook(w http.ResponseWriter, r *http.Request) {
	depthStr := r.URL.Query().Get("depth")
	depth := 20
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 && d <= 100 {
			depth = d
		}
	}

	book, err := s.engine.GetOrderBook("R.index", depth)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, book)
}

func (s *Server) handleGetMarketPositions(w http.ResponseWriter, r *http.Request) {
	positions := s.engine.GetAllPositions("R.index")
	respondJSON(w, http.StatusOK, positions)
}

func (s *Server) handleGetMarketOpenInterest(w http.ResponseWriter, r *http.Request) {
	oi := s.engine.GetOpenInterestBreakdown("R.index")
	respondJSON(w, http.StatusOK, oi)
}

func (s *Server) handleGetMarketTrades(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	trades := s.engine.GetRecentTrades("R.index", limit)
	respondJSON(w, http.StatusOK, trades)
}

func (s *Server) handleGetMarketLiquidations(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	liquidations := s.engine.GetRecentLiquidations("R.index", limit)
	respondJSON(w, http.StatusOK, liquidations)
}

func (s *Server) handleGetMarketStats(w http.ResponseWriter, r *http.Request) {
	stats := s.engine.GetMarketStats("R.index")
	respondJSON(w, http.StatusOK, stats)
}

// Auth handlers (simplified - no real auth for now)

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string            `json:"username"`
		Password string            `json:"password"`
		Type     domain.TraderType `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "username and password required")
		return
	}

	if req.Type == "" {
		req.Type = domain.TraderTypeHuman
	}

	trader := &domain.Trader{
		ID:        uuid.New(),
		Username:  req.Username,
		Type:      req.Type,
		Balance:   decimal.NewFromInt(10000), // Starting balance
		CreatedAt: time.Now(),
		TotalPnL:  decimal.Zero,
	}

	s.engine.RegisterTrader(trader)

	// Return trader with a simple token (trader ID as token for simplicity)
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"trader": trader,
		"token":  trader.ID.String(),
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Find trader by username
	traders := s.engine.GetAllTraders()
	for _, trader := range traders {
		if trader.Username == req.Username {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"trader": trader,
				"token":  trader.ID.String(),
			})
			return
		}
	}

	respondError(w, http.StatusUnauthorized, "invalid credentials")
}
