from Stack import Stack


bid = Stack()
ask = Stack(False)

bid.push(10,3)
bid.push(20,5)

ask.push(15,2)
ask.push(25,6)

bid.size()
print(bid)

ask.size()
print(ask)