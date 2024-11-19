from Stack import Stack
from Order import Order
from OrderBook import OrderBook

# Test Stack

bid = Stack()
ask = Stack(False)

bid.push(10,3)
bid.push(20,5)

ask.push(15,2)
ask.push(25,6)

bid.size()

ask.size()

order = Order(1, "buy", 10, 5)
print(order)

order2 = Order(2, "sell", 15, 3)

orderbook = OrderBook(bid, ask)
orderbook.add_order(order2)
print(orderbook)