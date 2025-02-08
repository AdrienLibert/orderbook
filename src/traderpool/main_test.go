package main

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTrader(t *testing.T) {
	trader := NewTrader("test_trader", 100.0, 10.0, nil)

	assert.Equal(t, "test_trader", trader.traderId)
	assert.Equal(t, 100.0, trader.limitPrice)
	assert.Equal(t, 10.0, trader.quantity)
	assert.NotNil(t, trader)
}

func TestComputeThetaEstimate(t *testing.T) {
	tests := []struct {
		name             string
		theta            float64
		limitPrice       float64
		equilibriumPrice float64
		quantity         float64
		expectedMin      float64
		expectedMax      float64
	}{
		{"Basic positive", 0.5, 100.0, 95.0, 10.0, -2.0, 2.0},
		{"Basic negative", -0.8, 100.0, 105.0, -10.0, -2.0, 2.0},
		{"Zero theta", 0.0, 100.0, 100.0, 5.0, -2.0, 2.0},
		{"Very large theta", 5.0, 120.0, 80.0, 15.0, -2.0, 2.0},
		{"Very small theta", -5.0, 90.0, 110.0, -20.0, -2.0, 2.0},
		{"Edge case: theta = 1", 1.0, 110.0, 100.0, 8.0, -2.0, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trader := Trader{
				theta:            tt.theta,
				limitPrice:       tt.limitPrice,
				equilibriumPrice: tt.equilibriumPrice,
				quantity:         tt.quantity,
			}

			result := trader.computeThetaEstimate()

			assert.True(t, result >= tt.expectedMin && result <= tt.expectedMax, "Result %.6f should be within range [%.6f, %.6f]", result, tt.expectedMin, tt.expectedMax)

			assert.False(t, math.IsNaN(result), "Result should not be NaN")
			assert.False(t, math.IsInf(result, 0), "Result should not be Infinite")
		})
	}
}

func TestCalculateTarget(t *testing.T) {
	tests := []struct {
		name             string
		limitPrice       float64
		equilibriumPrice float64
		aggressiveness   float64
		theta            float64
		isBuy            bool
		expectedMin      float64
		expectedMax      float64
	}{
		{"Basic Buy Positive Aggression", 95.0, 100.0, 0.5, 1.0, true, 90.0, 100.0},
		{"Basic Sell Positive Aggression", 105.0, 100.0, 0.5, 1.0, false, 100.0, 110.0},
		{"Basic Buy Negative Aggression", 95.0, 100.0, -0.5, 1.0, true, 80.0, 95.0},
		{"Basic Sell Negative Aggression", 105.0, 100.0, -0.5, 1.0, false, 105.0, 120.0},
		{"Limit Price Greater than Equilibrium Buy", 110.0, 100.0, 0.5, 1.0, true, 100.0, 110.0},
		{"Limit Price Greater than Equilibrium Sell", 110.0, 100.0, 0.5, 1.0, false, 105.0, 115.0},
		{"Edge Case: Zero Aggression Buy", 100.0, 100.0, 0.0, 1.0, true, 100.0, 100.0},
		{"Edge Case: Zero Aggression Sell", 100.0, 100.0, 0.0, 1.0, false, 100.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trader := Trader{}
			result := trader.calculateTarget(tt.limitPrice, tt.equilibriumPrice, tt.aggressiveness, tt.theta, tt.isBuy)

			assert.True(t, result >= tt.expectedMin && result <= tt.expectedMax, "Result %.6f should be within range [%.6f, %.6f]", result, tt.expectedMin, tt.expectedMax)
		})
	}
}

func TestPublishOrder(t *testing.T) {
	order := publishOrder("test_trader", 10.0, 101.5, "buy")

	assert.NotNil(t, order)
	assert.Equal(t, "buy", order.OrderType)
	assert.Equal(t, 10.0, order.Quantity)
	assert.Equal(t, 101.5, order.Price)
}

func TestConvertOrderToMessage(t *testing.T) {
	order := Order{
		OrderID:   "123",
		OrderType: "buy",
		Price:     100.0,
		Quantity:  10.0,
		Timestamp: 1645454,
	}
	message := convertOrderToMessage(order)

	var decoded Order
	err := json.Unmarshal(message, &decoded)
	assert.Nil(t, err)
	assert.Equal(t, order, decoded)
}

func TestConvertMessageToTrade(t *testing.T) {
	trade := Trade{
		OrderId:  "123",
		Quantity: 10.0,
		Price:    100.0,
		Action:   "buy",
		Status:   "open",
	}

	message, _ := json.Marshal(trade)
	decodedTrade, err := convertMessageToTrade(message)

	assert.Nil(t, err)
	assert.Equal(t, trade, decodedTrade)
}
