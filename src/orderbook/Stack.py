class Stack:
    def __init__(self,is_bid = True):
        self.order = []
        self.is_bid = is_bid
    def is_empty(self):
        return len(self.order) == 0
    
    def sort(self):
        self.order.sort(key = lambda x: x['Price'], reverse = self.is_bid)
    def push(self, price, quantity):
        for i in range(len(self.order)):
            if self.order[i]['Price'] == price:
                self.order[i]['Quantity'] += quantity
                return
        self.order.append({'Price' : price, 'Quantity' : quantity})
        self.sort()
    
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
    
    def to_dict(self):
        return {
            "order": self.order
        }
    
    def __repr__(self):
        return json.dumps(self.to_dict())