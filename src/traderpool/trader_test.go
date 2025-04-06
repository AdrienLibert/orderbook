package main

import (
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func TestTraderPool(t *testing.T) {
	// Seed random number generator for consistent results
	rand.Seed(time.Now().UnixNano())

	sharedMidPrice := &MidPrice{Value: 100.0}
	orderChan := make(chan Order, 20)
	kc := NewMockKafkaClient(t)

	// Create pool of 4 traders
	traders := []*Trader{
		{
			MidPrice:           sharedMidPrice,
			AggressivenessBuy:  2.0,
			AggressivenessSell: 1.5,
			Sign:               1,
			TraderId:           "T1",
			KafkaClient:        kc.KafkaClient,
		},
		{
			MidPrice:           sharedMidPrice,
			AggressivenessBuy:  1.5,
			AggressivenessSell: 2.0,
			Sign:               -1,
			TraderId:           "T2",
			KafkaClient:        kc.KafkaClient,
		},
		{
			MidPrice:           sharedMidPrice,
			AggressivenessBuy:  1.0,
			AggressivenessSell: 1.0,
			Sign:               1,
			TraderId:           "T3",
			KafkaClient:        kc.KafkaClient,
		},
		{
			MidPrice:           sharedMidPrice,
			AggressivenessBuy:  2.5,
			AggressivenessSell: 2.5,
			Sign:               -1,
			TraderId:           "T4",
			KafkaClient:        kc.KafkaClient,
		},
	}

	t.Run("InitialTargetPrices", func(t *testing.T) {
		for _, trader := range traders {
			trader.TargetPrices()
			expectedBuy := trader.MidPrice.Value + (trader.AggressivenessBuy * 0.5)
			expectedSell := trader.MidPrice.Value - (trader.AggressivenessSell * 0.5)

			if trader.TargetBuy != expectedBuy {
				t.Errorf("Trader %s: Expected TargetBuy %.2f, got %.2f",
					trader.TraderId, expectedBuy, trader.TargetBuy)
			}
			if trader.TargetSell != expectedSell {
				t.Errorf("Trader %s: Expected TargetSell %.2f, got %.2f",
					trader.TraderId, expectedSell, trader.TargetSell)
			}
		}
	})

	t.Run("ConcurrentTrading", func(t *testing.T) {
		var wg sync.WaitGroup
		ordersReceived := make([]Order, 0, 20)
		var mu sync.Mutex

		for _, trader := range traders {
			wg.Add(1)
			go func(tr *Trader) {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					order := tr.Trade(orderChan)
					mu.Lock()
					ordersReceived = append(ordersReceived, order)
					mu.Unlock()
					time.Sleep(time.Millisecond * 10)
				}
			}(trader)
		}

		go func() {
			for order := range orderChan {
				mu.Lock()
				ordersReceived = append(ordersReceived, order)
				mu.Unlock()
			}
		}()

		wg.Wait()
		close(orderChan)

		if len(ordersReceived) < 20 {
			t.Errorf("Expected at least 20 orders, got %d", len(ordersReceived))
		}

		for _, order := range ordersReceived {
			if order.Quantity == 0 {
				t.Errorf("Received empty order: %+v", order)
				continue
			}
			if order.Quantity > 50 || order.Quantity < -50 {
				t.Errorf("Order quantity out of range: %d", order.Quantity)
			}
		}
	})

	t.Run("PoolStartStop", func(t *testing.T) {
		var wg sync.WaitGroup
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		for _, trader := range traders {
			wg.Add(1)
			go func(tr *Trader) {
				defer wg.Done()
				go tr.Start()
			}(trader)
		}

		signals <- os.Interrupt
		wg.Wait()
	})

	t.Run("PriceUpdates", func(t *testing.T) {
		newPrice := 105.0
		sharedMidPrice.Value = newPrice

		for _, trader := range traders {
			trader.TargetPrices()
			expectedBuy := newPrice + (trader.AggressivenessBuy * 0.5)
			expectedSell := newPrice - (trader.AggressivenessSell * 0.5)

			if trader.TargetBuy != expectedBuy {
				t.Errorf("Trader %s: Expected TargetBuy %.2f, got %.2f",
					trader.TraderId, expectedBuy, trader.TargetBuy)
			}
			if trader.TargetSell != expectedSell {
				t.Errorf("Trader %s: Expected TargetSell %.2f, got %.2f",
					trader.TraderId, expectedSell, trader.TargetSell)
			}
		}
	})

	t.Run("TradeConsumption", func(t *testing.T) {
		kc := NewMockKafkaClient(t)
		trader := &Trader{
			MidPrice:           sharedMidPrice,
			AggressivenessBuy:  2.0,
			AggressivenessSell: 1.5,
			Sign:               1,
			TraderId:           "T1",
			KafkaClient:        kc.KafkaClient,
		}

		consumer := make(MockConsumerChan, 10)
		errors := make(chan *sarama.ConsumerError, 10)
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		var consumedCount int
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				select {
				case msg := <-consumer:
					trade, err := convertMessageToTrade(msg.Value)
					if err != nil {
						t.Errorf("Error converting message: %v", err)
						return
					}
					if trade.OrderId == trader.TraderId {
						consumedCount++
						switch trade.Status {
						case "closed":
							if rand.Float64() < 0.75 {
								trader.Sign = -1 * trader.Sign
							}
						default:
							t.Logf("Processing non-closed trade, initial sign: %d", trader.Sign)
							trader.AggressivenessSell *= 1.4
							trader.AggressivenessBuy *= 1.4
							trader.Sign = -1 * trader.Sign
							t.Logf("After processing: sign: %d, buy agg: %.2f, sell agg: %.2f",
								trader.Sign, trader.AggressivenessBuy, trader.AggressivenessSell)
						}
					}
				case <-signals:
					return
				case <-errors:
					consumedCount++
					return
				}
			}
		}()

		// Test closed trade
		initialSign := trader.Sign
		t.Logf("Testing closed trade, initial sign: %d", initialSign)
		consumer.SendOrder("T1", "closed")
		time.Sleep(50 * time.Millisecond)

		if consumedCount != 1 {
			t.Errorf("Expected consumedCount to be 1, got %d", consumedCount)
		}
		if trader.Sign != initialSign && trader.Sign != -initialSign {
			t.Errorf("Unexpected sign value after closed trade: %d (initial: %d)", trader.Sign, initialSign)
		}

		// Test non-closed trade
		initialBuyAgg := trader.AggressivenessBuy
		initialSellAgg := trader.AggressivenessSell
		initialSign = trader.Sign
		t.Logf("Testing non-closed trade, initial sign: %d", initialSign)
		consumer.SendOrder("T1", "open")
		time.Sleep(50 * time.Millisecond)

		if consumedCount != 2 {
			t.Errorf("Expected consumedCount to be 2, got %d", consumedCount)
		}
		if trader.Sign != -initialSign {
			t.Errorf("Expected sign to flip after non-closed trade, got %d (initial: %d)", trader.Sign, initialSign)
		}
		if trader.AggressivenessBuy != initialBuyAgg*1.4 {
			t.Errorf("Expected AggressivenessBuy %.2f, got %.2f",
				initialBuyAgg*1.4, trader.AggressivenessBuy)
		}
		if trader.AggressivenessSell != initialSellAgg*1.4 {
			t.Errorf("Expected AggressivenessSell %.2f, got %.2f",
				initialSellAgg*1.4, trader.AggressivenessSell)
		}

		t.Logf("Testing unrelated trade, current sign: %d", trader.Sign)
		consumer.SendOrder("T2", "closed")
		time.Sleep(50 * time.Millisecond)
		if consumedCount != 2 {
			t.Errorf("Expected consumedCount to remain 2, got %d", consumedCount)
		}

		signals <- os.Interrupt
		wg.Wait()
	})
}
