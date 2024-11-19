# Order to create an order object
class Order:
    def __init__(self, order_id, order_type, price, quantity):
        self.order_id = order_id
        self.order_type = order_type
        self.price = price
        self.quantity = quantity
    
    def get_price(self):
        return self.price
    
    def get_quantity(self):
        return self.quantity

    def set_quantity(self, quantity):
        self.quantity = quantity
    
    def set_price(self, price):
        self.price = price

    def __str__(self):
        return f"Order ID: {self.order_id}, Order Type: {self.order_type}, Price: {self.price}, Quantity: {self.quantity}"