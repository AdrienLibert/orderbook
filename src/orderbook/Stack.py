# Stack class for bid and ask
class Stack:
    def __init__(self,is_bid = True):
        self.order = []
        self.is_bid = is_bid # True for bid, False for ask because bid is in descending order and ask is in ascending order

    def is_empty(self):
        return len(self.order) == 0

    def push(self, price, quantity):
        self.order.append({'Price' : price, 'Quantity' : quantity})
        self.order.sort(key = lambda x: x['Price'], reverse = self.is_bid)

    def pop(self):
        if self.is_empty():
            return None
        return self.order.pop(0)

    def peek(self):
        if self.is_empty():
            return None
        return self.order[0]
    def size(self):
        return len(self.order)
    
    def __str__(self):
        return str(self.order)