# Order to create an order object
class Order:
    def __init__(self, order_id, price, quantity,order_type):
        self.order_id = order_id
        self.price = price
        self.quantity = quantity
        self.order_type = order_type

    def get_price(self):
        return self.price
    
    def get_quantity(self):
        return self.quantity

    def get_order_id(self):
        return self.order_id
    
    def get_order_type(self):
        return self.order_type

    def set_quantity(self, quantity):
        self.quantity = quantity
    
    def set_price(self, price):
        self.price = price

    def __str__(self):
        return f"Order ID: {self.order_id}, Order Type: {self.order_type}, Price: {self.price}, Quantity: {self.quantity}"

# Subclass for buy orders
# class OrderBuy(Order):
#     def __init__(self, order_id, price, quantity):
#         super().__init__(order_id, price, quantity, "BUY")


# # Subclass for sell orders
# class OrderSell(Order):
#     def __init__(self, order_id, price, quantity):
#         super().__init__(order_id, price, quantity, "SELL")