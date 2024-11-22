from Stack import Stack
from Order import Order


class OrderBook:
    def __init__(self, bid, ask):

        self.bid = bid  # Stack for bids
        self.ask = ask  # Stack for asks

    def get_bid(self):
        return self.bid

    def get_ask(self):
        return self.ask

    def _get_best_bid(self):
        return self.bid.peek() if not self.bid.is_empty() else None

    def _get_best_ask(self):
        return self.ask.peek() if not self.ask.is_empty() else None

    def add_order(self, order):

        if order.order_type == "buy":
            self.match_buy_order(order)
        elif order.order_type == "sell":
            self.match_sell_order(order)
        else:
            raise ValueError("Invalid order type")
    
        print("OrderBook state:")
        print(f"Bid: {self.bid}")
        print(f"Ask: {self.ask}")

    def match_buy_order(self, order):

        while order.get_quantity() > 0 and not self.ask.is_empty() and order.get_price() >= self.ask.peek()["Price"]:
            best_ask = self.ask.peek()
            trade_quantity = min(order.get_quantity(), best_ask["Quantity"])

            print(f"Trade: Buy {trade_quantity} @ {best_ask['Price']}")

            order.set_quantity(order.get_quantity() - trade_quantity)
            best_ask["Quantity"] -= trade_quantity

            if best_ask["Quantity"] == 0:
                self.ask.pop()

        if order.get_quantity() > 0:
            self.bid.push({"Price": order.get_price(), "Quantity": order.get_quantity()})

    def match_sell_order(self, order):

        while order.get_quantity() > 0 and not self.bid.is_empty() and order.get_price() <= self.bid.peek()["Price"]:
            best_bid = self.bid.peek()
            trade_quantity = min(order.get_quantity(), best_bid["Quantity"])

            print(f"Trade: Sell {trade_quantity} @ {best_bid['Price']}")

            order.set_quantity(order.get_quantity() - trade_quantity)
            best_bid["Quantity"] -= trade_quantity

            if best_bid["Quantity"] == 0:
                self.bid.pop()

        if order.get_quantity() > 0:
            self.ask.push({"Price": order.get_price(), "Quantity": order.get_quantity()})

    def __str__(self):
        return f"Bid: {self.bid}, Ask: {self.ask}"
